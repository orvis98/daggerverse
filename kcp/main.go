// Runs a kcp server than can be accessed both locally and in your pipelines

package main

import (
	"context"
	"dagger/kcp/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/kops/pkg/kubeconfig"
)

type Kcp struct {
	// The kcp version.
	Version string
	// The generated client CA certificate.
	ClientCaCertificate *dagger.File
	// The generated client CA private key.
	ClientCaKey *dagger.File
	// The generated client certificate for kcp-admin.
	ClientCertificate *dagger.File
	// The generated client certificate private key for kcp-admin.
	ClientCertificateKey *dagger.File
}

func New(
	// The kcp version to use.
	// +optional
	// +default="0.28.1"
	version string,
	// The openssl image to use for generating certificates.
	// +optional
	// +default="alpine/openssl:latest"
	opensslImage string,
) *Kcp {
	ctr := dag.Container().
		From(opensslImage).
		// Generate CA certificate
		WithExec([]string{"openssl", "req", "-x509", "-noenc", "-days", "365", "-newkey", "rsa:2048",
			"-subj", "/CN=kcp-ca", "-out", "ca.crt", "-keyout", "ca.key"}).
		// Generate client certificate for kcp-admin
		WithExec([]string{"openssl", "req", "-noenc", "-days", "30", "-newkey", "rsa:2048",
			"-subj", "/O=system:kcp:admin/CN=kcp-admin",
			"-out", "admin.csr", "-keyout", "admin.key"}).
		WithExec([]string{"openssl", "x509", "-req", "-in", "admin.csr", "-copy_extensions", "copyall",
			"-CA", "ca.crt", "-CAkey", "ca.key", "-CAcreateserial", "-out", "admin.crt"})
	return &Kcp{
		Version:              version,
		ClientCaCertificate:  ctr.File("ca.crt"),
		ClientCaKey:          ctr.File("ca.key"),
		ClientCertificate:    ctr.File("admin.crt"),
		ClientCertificateKey: ctr.File("admin.key"),
	}
}

// Returns a kcp server container.
func (m *Kcp) Server() *dagger.Container {
	return dag.Container().
		From(fmt.Sprintf("ghcr.io/kcp-dev/kcp:v%s", m.Version)).
		WithDirectory(".kcp", dag.Directory(), dagger.ContainerWithDirectoryOpts{Owner: "65532:65532"}).
		WithEnvVariable("CACHEBUST", time.Now().String()).
		WithFile("ca.crt", m.ClientCaCertificate, dagger.ContainerWithFileOpts{Owner: "65532:65532"}).
		WithEntrypoint([]string{"/kcp", "start", "--client-ca-file=ca.crt"}).
		WithExposedPort(6443, dagger.ContainerWithExposedPortOpts{Description: "kcp api-server"})
}

// Returns the kcp-admin kubeconfig.
func (m *Kcp) Config(ctx context.Context) (*dagger.File, error) {
	svc, err := m.Server().AsService().Start(ctx)
	if err != nil {
		return nil, err
	}
	ep, err := svc.Endpoint(ctx, dagger.ServiceEndpointOpts{Scheme: "https", Port: 6443})
	if err != nil {
		return nil, err
	}
	crt, err := m.ClientCertificate.Contents(ctx)
	if err != nil {
		return nil, err
	}
	key, err := m.ClientCertificateKey.Contents(ctx)
	if err != nil {
		return nil, err
	}
	cfg, err := json.Marshal(kubeconfig.KubectlConfig{
		ApiVersion: "v1",
		Kind:       "Config",
		Clusters: []*kubeconfig.KubectlClusterWithName{
			{
				Name: "root",
				Cluster: kubeconfig.KubectlCluster{
					Server: fmt.Sprintf("%s/clusters/root", ep),
				},
			},
		},
		Users: []*kubeconfig.KubectlUserWithName{
			{
				Name: "admin",
				User: kubeconfig.KubectlUser{
					ClientCertificateData: []byte(crt),
					ClientKeyData:         []byte(key),
				},
			},
		},
		Contexts: []*kubeconfig.KubectlContextWithName{
			{
				Name: "current",
				Context: kubeconfig.KubectlContext{
					Cluster: "root",
					User:    "admin",
				},
			},
		},
		CurrentContext: "current",
	})
	if err != nil {
		return nil, err
	}
	hack, err := dag.Container().
		From("alpine").
		WithExec([]string{"apk", "add", "jq", "curl"}).
		WithExec([]string{"sh", "-c", `until [ "$(curl --insecure -s -o /dev/null -w "%{http_code}" ` + ep + `/healthz)" = "200" ]; do sleep 1; done;`}).
		WithExec([]string{"jq", `.clusters[0].cluster["insecure-skip-tls-verify"] = true`}, dagger.ContainerWithExecOpts{Stdin: string(cfg)}).
		Stdout(ctx)
	if err != nil {
		return nil, err
	}
	return dag.File("config", hack), nil
}

// Returns a kcp client container with the kcp plugins and the kcp-admin kubeconfig.
func (m *Kcp) Client(
	ctx context.Context,
	// The system platform (should always be "linux").
	// +optional
	// +default="linux"
	platform string,
	// The system architecture (should be "amd64" or "arm64")
	// +optional
	// +default="amd64"
	arch string,
) (*dagger.Container, error) {
	cfg, err := m.Config(ctx)
	if err != nil {
		return nil, err
	}
	return dag.Container().
		From("alpine/kubectl:latest").
		WithExec([]string{"wget", fmt.Sprintf("https://github.com/kcp-dev/kcp/releases/download/v%s/kubectl-create-workspace-plugin_%s_%s_%s.tar.gz", m.Version, m.Version, platform, arch)}).
		WithExec([]string{"tar", "-xf", fmt.Sprintf("kubectl-create-workspace-plugin_%s_%s_%s.tar.gz", m.Version, platform, arch), "-C", "/usr/local/bin", "--strip-components=1"}).
		WithExec([]string{"wget", fmt.Sprintf("https://github.com/kcp-dev/kcp/releases/download/v%s/kubectl-kcp-plugin_%s_%s_%s.tar.gz", m.Version, m.Version, platform, arch)}).
		WithExec([]string{"tar", "-xf", fmt.Sprintf("kubectl-kcp-plugin_%s_%s_%s.tar.gz", m.Version, platform, arch), "-C", "/usr/local/bin", "--strip-components=1"}).
		WithExec([]string{"wget", fmt.Sprintf("https://github.com/kcp-dev/kcp/releases/download/v%s/kubectl-ws-plugin_%s_%s_%s.tar.gz", m.Version, m.Version, platform, arch)}).
		WithExec([]string{"tar", "-xf", fmt.Sprintf("kubectl-ws-plugin_%s_%s_%s.tar.gz", m.Version, platform, arch), "-C", "/usr/local/bin", "--strip-components=1"}).
		WithFile("/.kube/config", cfg).
		WithEnvVariable("KUBECONFIG", "/.kube/config"), nil
}

func setupWorkspace(
	ctx context.Context,
	ctr *dagger.Container,
	path string,
	workspace *dagger.Directory,
) (*dagger.Container, error) {
	ctr = ctr.WithExec([]string{"kubectl", "ws", fmt.Sprintf(":%s", path)}).
		WithDirectory("/build", workspace).
		WithWorkdir("/build")
	ents, err := workspace.Entries(ctx)
	if err != nil {
		return nil, err
	}
	for _, ent := range ents {
		if strings.HasSuffix(ent, ".yaml") || strings.HasSuffix(ent, ".yml") {
			ctr = ctr.WithExec([]string{"kubectl", "apply", "-f", ent})
		} else if strings.HasSuffix(ent, "/") {
			cut := strings.TrimSuffix(ent, "/")
			ctr = ctr.WithExec([]string{"kubectl", "create", "workspace", cut})
			ws, err := setupWorkspace(ctx, ctr, fmt.Sprintf("%s:%s", path, cut), workspace.Directory(cut))
			if err != nil {
				return nil, err
			}
			ec, err := ws.ExitCode(ctx)
			if err != nil {
				return nil, err
			}
			ctr = ctr.WithExec([]string{"echo", fmt.Sprintf("configured %s with exit code %d", cut, ec)})
		}
	}
	return ctr, nil
}

func (m *Kcp) WithWorkspaces(
	ctx context.Context,
	// The system platform (should always be "linux").
	// +optional
	// +default="linux"
	platform string,
	// The system architecture (should be "amd64" or "arm64")
	// +optional
	// +default="amd64"
	arch string,
	// The directory tree to bootstrap in kcp.
	// +optional
	root *dagger.Directory,
) (*dagger.Container, error) {
	ctr, err := m.Client(ctx, platform, arch)
	if err != nil {
		return nil, err
	}
	ctr = ctr.WithDirectory("/build", root).
		WithWorkdir("/build")
	ctr, err = setupWorkspace(ctx, ctr, "root", root)
	if err != nil {
		return nil, err
	}
	return ctr, nil
}
