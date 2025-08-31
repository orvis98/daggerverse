// Runs a kcp server than can be accessed both locally and in your pipelines

package main

import (
	"context"
	"dagger/kcp/internal/dagger"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/kops/pkg/kubeconfig"
)

type Kcp struct {
	// +private
	Name string

	// Returns the kcp-admin kubeconfig as a file.
	Kubeconfig *dagger.File

	// Returns the kcp server container as a service.
	Server *dagger.Service
}

type PKI struct {
	CACertificate          *dagger.File
	ServerCertificate      *dagger.File
	ServerPrivateKey       *dagger.File
	AdminClientCertificate *dagger.File
	AdminPrivateKey        *dagger.File
}

func bootstrapPKI() *PKI {
	ctr := dag.Container().
		From("alpine").
		WithExec([]string{"apk", "add", "openssl"}).
		// Generate CA certificate
		WithExec([]string{"openssl", "req", "-x509", "-noenc", "-days", "365", "-newkey", "rsa:2048",
			"-subj", "/CN=kcp-ca", "-out", "ca.crt", "-keyout", "ca.key"}).
		// Generate server certificate
		WithExec([]string{"openssl", "req", "-noenc", "-days", "90", "-newkey", "rsa:2048",
			"-subj", "/CN=kcp-server", "-addext", "subjectAltName=DNS:kcp", "-out=server.csr", "-keyout=server.key"}).
		WithExec([]string{"openssl", "x509", "-req", "-in", "server.csr", "-copy_extensions", "copyall",
			"-CA", "ca.crt", "-CAkey", "ca.key", "-CAcreateserial", "-out", "server.crt"}).
		// Generate client certificate for kcp-admin
		WithExec([]string{"openssl", "req", "-noenc", "-days", "30", "-newkey", "rsa:2048",
			"-subj", "/O=system:kcp:admin/CN=kcp-admin",
			"-out", "admin.csr", "-keyout", "admin.key"}).
		WithExec([]string{"openssl", "x509", "-req", "-in", "admin.csr", "-copy_extensions", "copyall",
			"-CA", "ca.crt", "-CAkey", "ca.key", "-CAcreateserial", "-out", "admin.crt"})
	return &PKI{
		CACertificate:          ctr.File("ca.crt"),
		ServerCertificate:      ctr.File("server.crt"),
		ServerPrivateKey:       ctr.File("server.key"),
		AdminClientCertificate: ctr.File("admin.crt"),
		AdminPrivateKey:        ctr.File("admin.key"),
	}
}

func setupAdminKubeconfig(ctx context.Context, pki *PKI) (*dagger.File, error) {
	caCert, err := pki.CACertificate.Contents(ctx)
	if err != nil {
		return nil, err
	}
	clientCert, _ := pki.AdminClientCertificate.Contents(ctx)
	if err != nil {
		return nil, err
	}
	clientKey, _ := pki.AdminPrivateKey.Contents(ctx)
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
					Server:                   "https://kcp:6443/clusters/root",
					CertificateAuthorityData: []byte(caCert),
				},
			},
		},
		Users: []*kubeconfig.KubectlUserWithName{
			{
				Name: "admin",
				User: kubeconfig.KubectlUser{
					ClientCertificateData: []byte(clientCert),
					ClientKeyData:         []byte(clientKey),
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
	return dag.File("admin.kubeconfig", string(cfg)), err
}

func New(
	ctx context.Context,
	// When specified a cache volume will be configured.
	// +optional
	name string,
	// The kcp server container image.
	// +optional
	// +default="ghcr.io/kcp-dev/kcp:v0.27.0"
	image string,
) *Kcp {
	pki := bootstrapPKI()
	cfg, err := setupAdminKubeconfig(ctx, pki)
	if err != nil {
		return nil
	}
	return &Kcp{
		Kubeconfig: cfg,
		Server: dag.Container().
			From(image).
			With(func(c *dagger.Container) *dagger.Container {
				// When a name is set, the caller is responsible for health checking the server after restarts.
				if name != "" {
					c = c.WithMountedCache(".kcp", dag.CacheVolume("kcp_data_"+name), dagger.ContainerWithMountedCacheOpts{Owner: "65532:65532"})
				} else {
					c = c.WithDirectory(".kcp", dag.Directory(), dagger.ContainerWithDirectoryOpts{Owner: "65532:65532"}).
						WithEnvVariable("CACHEBUST", time.Now().String())
				}
				return c
			}).
			WithFile("ca.crt", pki.CACertificate, dagger.ContainerWithFileOpts{Owner: "65532:65532"}).
			WithFile("server.crt", pki.ServerCertificate, dagger.ContainerWithFileOpts{Owner: "65532:65532"}).
			WithFile("server.key", pki.ServerPrivateKey, dagger.ContainerWithFileOpts{Owner: "65532:65532"}).
			WithEntrypoint([]string{"/kcp", "start", "--client-ca-file=ca.crt",
				"--tls-cert-file=server.crt", "--tls-private-key-file=server.key"}).
			WithExposedPort(6443, dagger.ContainerWithExposedPortOpts{Description: "kcp api-server"}).
			AsService(),
	}
}

// Returns a container with the kcp kubectl plugins and a kcp server service binding.
func (m *Kcp) Kubectl(
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
	binDir := dag.Container().
		From("alpine/kubectl").
		WithExec([]string{"wget", fmt.Sprintf("https://github.com/kcp-dev/kcp/releases/download/v0.27.0/kubectl-create-workspace-plugin_0.27.0_%s_%s.tar.gz", platform, arch)}).
		WithExec([]string{"tar", "-xf", fmt.Sprintf("kubectl-create-workspace-plugin_0.27.0_%s_%s.tar.gz", platform, arch), "-C", "/usr/local/bin", "--strip-components=1"}).
		WithExec([]string{"wget", fmt.Sprintf("https://github.com/kcp-dev/kcp/releases/download/v0.27.0/kubectl-kcp-plugin_0.27.0_%s_%s.tar.gz", platform, arch)}).
		WithExec([]string{"tar", "-xf", fmt.Sprintf("kubectl-kcp-plugin_0.27.0_%s_%s.tar.gz", platform, arch), "-C", "/usr/local/bin", "--strip-components=1"}).
		WithExec([]string{"wget", fmt.Sprintf("https://github.com/kcp-dev/kcp/releases/download/v0.27.0/kubectl-ws-plugin_0.27.0_%s_%s.tar.gz", platform, arch)}).
		WithExec([]string{"tar", "-xf", fmt.Sprintf("kubectl-ws-plugin_0.27.0_%s_%s.tar.gz", platform, arch), "-C", "/usr/local/bin", "--strip-components=1"}).
		Directory("/usr/local/bin")
	return dag.Container().
		From("alpine/curl").
		WithServiceBinding("kcp", m.Server).
		WithExec([]string{"sh", "-c", `until [ "$(curl --insecure -s -o /dev/null -w "%{http_code}" https://kcp:6443/healthz)" = "200" ]; do sleep 1; done;`}).
		WithDirectory("/usr/local/bin", binDir).
		WithFile("/.kube/config", m.Kubeconfig).
		WithEnvVariable("KUBECONFIG", "/.kube/config"), nil
}
