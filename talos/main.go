// A generated module for Talos functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"dagger/talos/internal/dagger"
	"fmt"
)

const envoyConfig = `
---
static_resources:
  listeners:
  - address:
      socket_address:
        address: 0.0.0.0
        port_value: 6443
    filter_chains:
    - filters:
      - name: envoy.filters.network.tcp_proxy
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: kube-api
          stat_prefix: https_passthrough
  - address:
      socket_address:
        address: 0.0.0.0
        port_value: 50000
    filter_chains:
    - filters:
      - name: envoy.filters.network.tcp_proxy
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: talos-api
          stat_prefix: https_passthrough
  clusters:
  - name: kube-api
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: kube-api
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: %s
                port_value: 6443
  - name: talos-api
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: talos-api
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: %s
                port_value: 50000
`

func GetNode(name string, version string) *dagger.Container {
	return dag.Container().
		From(fmt.Sprintf("ghcr.io/siderolabs/talos:%s", version)).
		WithNewFile("/etc/hostname", name).
		WithMountedCache("/run", dag.CacheVolume(fmt.Sprintf("%s-run", name))).
		WithMountedCache("/system", dag.CacheVolume(fmt.Sprintf("%s-system", name))).
		WithMountedTemp("/tmp").
		WithMountedCache("/system/state", dag.CacheVolume(fmt.Sprintf("%s-system-state", name))).
		WithMountedCache("/var", dag.CacheVolume(fmt.Sprintf("%s-var", name))).
		WithMountedCache("/etc/cni", dag.CacheVolume(fmt.Sprintf("%s-etc-cni", name))).
		WithMountedCache("/etc/kubernetes", dag.CacheVolume(fmt.Sprintf("%s-etc-kubernetes", name))).
		WithMountedCache("/usr/libexec/kubernetes", dag.CacheVolume(fmt.Sprintf("%s-usr-libexec-kubernetes", name))).
		WithMountedCache("/opt", dag.CacheVolume(fmt.Sprintf("%s-opt", name))).
		WithEnvVariable("PLATFORM", "container").
		WithExec([]string{"/sbin/init"}, dagger.ContainerWithExecOpts{
			ExperimentalPrivilegedNesting: true,
			InsecureRootCapabilities:      true,
			NoInit:                        true,
		}).
		WithExposedPort(50000)
}

type Talos struct {
	// +private
	Name string
	// +private
	Version string
	// +private
	Controlplanes int
	// +private
	Workers int
	// +private
	VIP string
	// The Talos API credentials
	Talosconfig *dagger.File
	// The Kubernetes API credentials
	Kubeconfig *dagger.File
}

func New(
	// +optional
	// +default="talos-default"
	// the desired cluster name
	name string,
	// +optional
	// +default="v1.8.3"
	// the desired Talos version
	version string,
	// +optional
	// +default=1
	// the desired number of controlplanes
	controlplanes int,
	// +optional
	// +default=1
	// the desired number of workers
	workers int,
	// +optional
	// +default="10.87.13.37"
	// the desired cluster VIP
	vip string,
) (*Talos, error) {
	if controlplanes < 1 {
		return nil, fmt.Errorf("invalid number of controlplanes %d", controlplanes)
	} else if workers < 0 {
		return nil, fmt.Errorf("invalid number of workers %d", workers)
	}
	ctr := dag.Container().
		WithFile("/bin/talosctl", dag.Container().
			From(fmt.Sprintf("ghcr.io/siderolabs/talosctl:%s", version)).
			File("/talosctl"),
		).
		WithFile("/bin/wait4x", dag.Container().
			From("atkrad/wait4x").
			File("/usr/bin/wait4x")).
		WithFile("/bin/kubectl", dag.Container().
			From("bitnami/kubectl").
			File("/opt/bitnami/kubectl/bin/kubectl")).
		WithExec([]string{"talosctl", "gen", "config", name, fmt.Sprintf("https://%s:6443", vip), fmt.Sprintf("--additional-sans=localhost,%s", vip),
			fmt.Sprintf("--config-patch-control-plane=[{\"op\": \"add\", \"path\": \"/machine/network/interfaces\", \"value\": [{\"interface\": \"eth1\", \"dhcp\": true, \"vip\": {\"ip\": \"%s\"}}]}]", vip),
			"--with-docs=false", "--with-examples=false",
		}).
		WithEnvVariable("TALOSCONFIG", "talosconfig").
		WithExec([]string{"talosctl", "config", "endpoint", vip})
	for i := range controlplanes {
		hostname := fmt.Sprintf("%s-controlplane-%d", name, i+1)
		ctr = ctr.WithServiceBinding(hostname, GetNode(hostname, version).AsService()).
			WithExec([]string{"talosctl", "--talosconfig", "talosconfig", "-n", hostname, "apply", "--insecure", "-f", "controlplane.yaml",
				fmt.Sprintf("--config-patch=[{\"op\": \"add\", \"path\": \"/machine/network/hostname\", \"value\": \"%s\"}]", hostname)})
	}
	for i := range workers {
		hostname := fmt.Sprintf("%s-worker-%d", name, i+1)
		ctr = ctr.WithServiceBinding(hostname, GetNode(hostname, version).AsService()).
			WithExec([]string{"talosctl", "--talosconfig", "talosconfig", "-n", hostname, "apply", "--insecure", "-f", "worker.yaml",
				fmt.Sprintf("--config-patch=[{\"op\": \"add\", \"path\": \"/machine/network/hostname\", \"value\": \"%s\"}]", hostname)})
	}
	ctr = ctr.WithExec([]string{"talosctl", "-e", fmt.Sprintf("%s-controlplane-1", name), "-n", fmt.Sprintf("%s-controlplane-1", name), "bootstrap"}).
		WithExec([]string{"wait4x", "tcp", fmt.Sprintf("%s:6443", vip), "--timeout", "300s"}).
		WithExec([]string{"talosctl", "-n", fmt.Sprintf("%s-controlplane-1", name), "kubeconfig", "kubeconfig"}).
		WithEnvVariable("KUBECONFIG", "kubeconfig")
	for i := range controlplanes {
		hostname := fmt.Sprintf("%s-controlplane-%d", name, i+1)
		ctr = ctr.WithExec([]string{"kubectl", "wait", "--for=create", fmt.Sprintf("node/%s", hostname), "--timeout=300s"})
	}
	for i := range workers {
		hostname := fmt.Sprintf("%s-worker-%d", name, i+1)
		ctr = ctr.WithExec([]string{"kubectl", "wait", "--for=create", fmt.Sprintf("node/%s", hostname), "--timeout=300s"})
	}
	return &Talos{
		Name:          name,
		Version:       version,
		Controlplanes: controlplanes,
		Workers:       workers,
		VIP:           vip,
		Talosconfig:   ctr.File("talosconfig"),
		Kubeconfig:    ctr.File("kubeconfig"),
	}, nil
}

func (t *Talos) Proxy() *dagger.Service {
	ctr := dag.Container().
		From("envoyproxy/envoy:v1.32.1")
	for i := range t.Controlplanes {
		hostname := fmt.Sprintf("%s-controlplane-%d", t.Name, i+1)
		ctr = ctr.WithServiceBinding(hostname, GetNode(hostname, t.Version).WithExposedPort(6443).AsService())
	}
	for i := range t.Workers {
		hostname := fmt.Sprintf("%s-worker-%d", t.Name, i+1)
		ctr = ctr.WithServiceBinding(hostname, GetNode(hostname, t.Version).AsService())
	}
	return ctr.WithNewFile("envoy.yaml", fmt.Sprintf(envoyConfig, t.VIP, t.VIP)).
		WithExec([]string{"envoy", "-c", "envoy.yaml"}).
		WithExposedPort(6443).
		WithExposedPort(50000).
		AsService()
}
