// Kcp Go examples module

package main

import (
	"context"
	"dagger/examples/internal/dagger"
)

type Examples struct{}

// Creates and outputs an example Workspace tree.
func (m *Examples) PrintWorkspaceTree(ctx context.Context) (string, error) {
	return dag.Kcp(dagger.KcpOpts{Image: "ghcr.io/kcp-dev/kcp:v0.27.0"}).
		Kubectl(dagger.KcpKubectlOpts{Arch: "arm64"}).
		WithExec([]string{"kubectl", "create", "workspace", "org-1"}).
		WithExec([]string{"kubectl", "create", "workspace", "org-2"}).
		WithExec([]string{"kubectl", "ws", "tree"}).
		Stdout(ctx)
}
