// Runs a kcp server than can be accessed both locally and in your pipelines

package main

import (
	"dagger/kcp/internal/dagger"
	"fmt"
	"time"
)

type Kcp struct {
	// +private
	Name string

	// +private
	ConfigCache *dagger.CacheVolume

	// Returns the kcp server container.
	Container *dagger.Container
}

func New(
	// The name of this module instance (used for naming cache volumes).
	// +optional
	// +default="default"
	name string,
	// The kcp server container image.
	// +optional
	// +default="ghcr.io/kcp-dev/kcp:v0.28.1"
	image string,
) *Kcp {
	ccache := dag.CacheVolume("kcp_config_" + name)
	return &Kcp{
		Name:        name,
		ConfigCache: ccache,
		Container: dag.Container().
			From(image).
			WithMountedCache(".kcp", ccache, dagger.ContainerWithMountedCacheOpts{Owner: "65532:65532"}).
			// We bust the cache to have kcp regenerate certificates in case of IP address change.
			WithEnvVariable("CACHE", time.Now().String()).
			WithExec([]string{"rm", "-f", ".kcp/apiserver.crt", ".kcp/apiserver.key", ".kcp/admin.kubeconfig", ".kcp/.admin-token-store"}),
	}
}

// Returns the kcp server container as a service.
func (m *Kcp) Server() *dagger.Service {
	return m.Container.
		AsService()
}

// Returns the kcp server admin kubeconfig as a file.
func (m *Kcp) Config() *dagger.File {
	const interval = 0.5
	return dag.Container().
		From("alpine").
		WithMountedCache("/cache/kcp", m.ConfigCache, dagger.ContainerWithMountedCacheOpts{Owner: "65532:65532"}).
		// We need to bust the cache so we don't fetch the same file each time.
		WithEnvVariable("CACHE", time.Now().String()).
		WithExec([]string{"sh", "-c", `while [ ! -f "/cache/kcp/admin.kubeconfig" ]; do echo "admin.kubeconfig not ready, is server started?. waiting.. " && sleep ` + fmt.Sprintf("%.1f", interval) + `; done`}).
		WithExec([]string{"cp", "/cache/kcp/admin.kubeconfig", "config"}).
		File("config")
}
