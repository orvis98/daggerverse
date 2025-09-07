// Kcp Go examples module

package main

import (
	"context"
	"dagger/examples/internal/dagger"
)

type Examples struct{}

// Creates and outputs an example Workspace tree.
func (m *Examples) PrintWorkspaceTree(
	ctx context.Context,
	// +optional
	// +defaultPath="/root"
	workspaces *dagger.Directory,
) (string, error) {
	return dag.Kcp().
		WithWorkspaces(dagger.KcpWithWorkspacesOpts{Root: workspaces, Arch: "arm64"}).
		WithExec([]string{"kubectl", "ws", "tree"}).
		Stdout(ctx)
}
