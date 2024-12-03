// Runs Talos Linux in containers that can be accessed both locally and in your pipelines

package main

import (
	"context"
	"dagger/talos/internal/dagger"
	"fmt"
	"regexp"
)

type TalosNode struct {
	Hostname string
	Version  string
}

func NewNode(hostname string, version string) TalosNode {
	return TalosNode{
		Hostname: hostname,
		Version:  version,
	}
}

func (n *TalosNode) Service(withExposedAPI bool) *dagger.Service {
	// https://www.talos.dev/v1.8/talos-guides/install/local-platforms/docker/#running-talos-in-docker-manually
	ctr := dag.Container().
		From(fmt.Sprintf("ghcr.io/siderolabs/talos:%s", n.Version)).
		WithNewFile("/etc/hostname", n.Hostname).
		WithMountedCache("/run", dag.CacheVolume(fmt.Sprintf("%s-run", n.Hostname))).
		WithMountedCache("/system", dag.CacheVolume(fmt.Sprintf("%s-system", n.Hostname))).
		WithMountedTemp("/tmp").
		WithMountedCache("/system/state", dag.CacheVolume(fmt.Sprintf("%s-system-state", n.Hostname))).
		WithMountedCache("/var", dag.CacheVolume(fmt.Sprintf("%s-var", n.Hostname))).
		WithMountedCache("/etc/cni", dag.CacheVolume(fmt.Sprintf("%s-etc-cni", n.Hostname))).
		WithMountedCache("/etc/kubernetes", dag.CacheVolume(fmt.Sprintf("%s-etc-kubernetes", n.Hostname))).
		WithMountedCache("/usr/libexec/kubernetes", dag.CacheVolume(fmt.Sprintf("%s-usr-libexec-kubernetes", n.Hostname))).
		WithMountedCache("/opt", dag.CacheVolume(fmt.Sprintf("%s-opt", n.Hostname))).
		WithEnvVariable("PLATFORM", "container").
		WithExec([]string{"/sbin/init"}, dagger.ContainerWithExecOpts{
			ExperimentalPrivilegedNesting: true,
			InsecureRootCapabilities:      true,
			NoInit:                        true,
		}).
		WithExposedPort(50000)
	if withExposedAPI {
		ctr = ctr.WithExposedPort(6443)
	}
	return ctr.AsService().WithHostname(n.Hostname)
}

type Talos struct {
	// +private
	Name string
	// +private
	Version string
	// +private
	Controlplanes []TalosNode
	// +private
	Workers []TalosNode
}

func New(
	// +optional
	// +default="talos"
	// the desired cluster name
	name string,
	// +optional
	// +default="v1.8.3"
	// the desired Talos version
	version string,
	// +optional
	// +default=1
	// the desired number of controlplane nodes
	controlplanes int,
	// +optional
	// +default=1
	// the desired number of worker nodes
	workers int,
) *Talos {
	cps, ws := make([]TalosNode, controlplanes), make([]TalosNode, workers)
	for i := range controlplanes {
		cps[i] = NewNode(fmt.Sprintf("%s-controlplane-%d", name, i+1), version)
	}
	for i := range workers {
		ws[i] = NewNode(fmt.Sprintf("%s-worker-%d", name, i+1), version)
	}
	return &Talos{
		Name:          name,
		Version:       version,
		Controlplanes: cps,
		Workers:       ws,
	}
}

func (t *Talos) talosctlContainer() *dagger.Container {
	return dag.Container().
		WithFile("/bin/talosctl", dag.Container().
			From(fmt.Sprintf("ghcr.io/siderolabs/talosctl:%s", t.Version)).
			File("/talosctl")).
		WithExec([]string{"talosctl", "gen", "secrets"})
}

func (t *Talos) withTalosconfig(endpoint string, node string) *dagger.Container {
	return t.talosctlContainer().
		WithExec([]string{"talosctl", "gen", "config", t.Name, "https://talos:6443",
			"--with-secrets=secrets.yaml", "--output-types=talosconfig"}).
		WithEnvVariable("TALOSCONFIG", "talosconfig").
		WithExec([]string{"talosctl", "config", "endpoint", endpoint}).
		WithExec([]string{"talosctl", "config", "node", node})
}

// the generated Talos client configuration file
func (t *Talos) Talosconfig(
	// +optional
	// +default="localhost"
	// the endpoint to set in the configuration
	endpoint string,
	// +optional
	// +default="localhost"
	// the node to set in the configuration
	node string,
) *dagger.File {
	return t.withTalosconfig(endpoint, node).
		File("talosconfig")
}

func (t *Talos) withMachineConfig(
	ctx context.Context,
	// +optional
	// +default="10.87.13.37"
	// the desired cluster VIP
	vip string,
	// +optional
	// +default=[]
	// patch generated machineconfigs (applied to all node types)
	configPatch []string,
	// +optional
	// +default=[]
	// patch generated machineconfigs (applied to all node types)
	configPatchFile []*dagger.File,
) *dagger.Container {
	var flags []string
	for _, p := range configPatch {
		flags = append(flags, fmt.Sprintf("--config-patch=%s", p))
	}
	for _, p := range configPatchFile {
		n, _ := p.Name(ctx)
		flags = append(flags, fmt.Sprintf("--config-patch=@patches/%s", n))
	}
	return t.withTalosconfig(vip, t.Controlplanes[0].Hostname).
		WithFiles("patches", configPatchFile).
		WithExec(append([]string{"talosctl", "gen", "config", t.Name, fmt.Sprintf("https://%s:6443", vip), fmt.Sprintf("--additional-sans=localhost,talos"),
			fmt.Sprintf("--config-patch-control-plane=[{\"op\": \"add\", \"path\": \"/machine/network/interfaces\", \"value\": [{\"interface\": \"eth1\", \"dhcp\": true, \"vip\": {\"ip\": \"%s\"}}]}]", vip),
			"--with-secrets=secrets.yaml", "--with-docs=false", "--with-examples=false", "--output-types=controlplane,worker"}, flags...))
}

// bootstraps the etcd cluster and waits for first controlplane node to register
func (t *Talos) Bootstrap(
	ctx context.Context,
	// +optional
	// +default="10.87.13.37"
	// the desired cluster VIP
	vip string,
	// +optional
	// +default=[]
	// patch generated machineconfigs (applied to all node types)
	configPatch []string,
	// +optional
	// +default=[]
	// patch generated machineconfigs (applied to all node types)
	configPatchFile []*dagger.File,
) *dagger.Container {
	ctr := t.withMachineConfig(ctx, vip, configPatch, configPatchFile).
		WithFile("/bin/wait4x", dag.Container().
			From("atkrad/wait4x").
			File("/usr/bin/wait4x")).
		WithFile("/bin/kubectl", dag.Container().
			From("bitnami/kubectl").
			File("/opt/bitnami/kubectl/bin/kubectl"))
	for _, n := range t.Controlplanes {
		ctr = ctr.WithServiceBinding(n.Hostname, n.Service(false)).
			WithExec([]string{"talosctl", "--talosconfig", "talosconfig", "-n", n.Hostname, "apply", "--insecure", "-f", "controlplane.yaml",
				fmt.Sprintf("--config-patch=[{\"op\": \"add\", \"path\": \"/machine/network/hostname\", \"value\": \"%s\"}]", n.Hostname)})
	}
	for _, n := range t.Workers {
		ctr = ctr.WithServiceBinding(n.Hostname, n.Service(false)).
			WithExec([]string{"talosctl", "--talosconfig", "talosconfig", "-n", n.Hostname, "apply", "--insecure", "-f", "worker.yaml",
				fmt.Sprintf("--config-patch=[{\"op\": \"add\", \"path\": \"/machine/network/hostname\", \"value\": \"%s\"}]", n.Hostname)})
	}
	return ctr.WithExec([]string{"talosctl", "-e", t.Controlplanes[0].Hostname, "-n", t.Controlplanes[0].Hostname, "bootstrap"}).
		WithExec([]string{"wait4x", "tcp", fmt.Sprintf("%s:6443", vip), "--timeout", "300s"}).
		WithExec([]string{"talosctl", "-n", t.Controlplanes[0].Hostname, "kubeconfig", "kubeconfig"}).
		WithEnvVariable("KUBECONFIG", "kubeconfig").
		WithExec([]string{"kubectl", "wait", "--for=create", fmt.Sprintf("node/%s", t.Controlplanes[0].Hostname), "--timeout=300s"})
}

// returns a proxy service for the Talos controlplane
func (t *Talos) Proxy() *dagger.Service {
	cfg := `---
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
      endpoints:`
	for _, n := range t.Controlplanes {
		cfg = cfg + fmt.Sprintf(`
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: %s
                port_value: 6443`, n.Hostname)
	}
	cfg = cfg + `
  - name: talos-api
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: talos-api
      endpoints:`
	for _, n := range t.Controlplanes {
		cfg = cfg + fmt.Sprintf(`
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: %s
                port_value: 50000`, n.Hostname)
	}
	ctr := dag.Container().
		From("envoyproxy/envoy:v1.32.1").
		WithNewFile("config.yaml", cfg)
	for _, n := range t.Controlplanes {
		ctr = ctr.WithServiceBinding(n.Hostname, n.Service(true))
	}
	for _, n := range t.Workers {
		ctr = ctr.WithServiceBinding(n.Hostname, n.Service(false))
	}
	return ctr.WithExec([]string{"envoy", "-c", "config.yaml"}).
		WithExposedPort(6443).
		WithExposedPort(50000).
		AsService()
}

func (t *Talos) withKubeconfig(ctx context.Context, server string) *dagger.Container {
	ctr := t.withTalosconfig("talos", t.Controlplanes[0].Hostname).
		WithServiceBinding("talos", t.Proxy()).
		WithExec([]string{"talosctl", "kubeconfig", "kubeconfig"})
	kubeconfig, _ := ctr.File("kubeconfig").Contents(ctx)
	m := regexp.MustCompile(`server: https://(?:[0-9]{1,3}\.){3}[0-9]{1,3}:6443`)
	kubeconfig = m.ReplaceAllString(kubeconfig, fmt.Sprintf("server: %s", server))
	return ctr.WithNewFile("kubeconfig", kubeconfig).
		WithEnvVariable("KUBECONFIG", "kubeconfig")
}

// returns the admin kubeconfig for the cluster
func (t *Talos) Kubeconfig(
	ctx context.Context,
	// +optional
	// +default="https://localhost:6443"
	// the server to set in the configuration
	server string,
) *dagger.File {
	return t.withKubeconfig(ctx, server).
		File("kubeconfig")
}

// returns a container that can execute talosctl and kubectl commands
func (t *Talos) Container(ctx context.Context) *dagger.Container {
	return t.withKubeconfig(ctx, "https://talos:6443").
		WithFile("/bin/kubectl", dag.Container().
			From("bitnami/kubectl").
			File("/opt/bitnami/kubectl/bin/kubectl"))
}
