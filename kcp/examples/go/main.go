// Kcp Go examples module

package main

import (
	"context"
	"dagger/examples/internal/dagger"
)

type Examples struct{}

// Returns a container with kubectl configured for a healthy kcp server.
func (m *Examples) Kubectl(ctx context.Context) *dagger.Container {
	kcp := dag.Kcp(dagger.KcpOpts{Name: "examples86"})
	server, err := kcp.Server().Start(ctx)
	if err != nil {
		return nil
	}
	ep, err := server.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: 6443, Scheme: "https"})
	if err != nil {
		return nil
	}
	return dag.Container().
		// We wait for kcp's /healthz endpoint to respond with status 200.
		From("alpine/curl").
		WithExec([]string{"sh", "-c", `until [ "$(curl --insecure -s -o /dev/null -w "%{http_code}" ` + ep + `/healthz)" = "200" ]; do sleep 1; done;`}).
		// We setup the kubectl container.
		From("alpine/kubectl").
		WithFile("/.kube/config", kcp.Config()).
		WithEnvVariable("KUBECONFIG", "/.kube/config")
}

// Display the WorkspaceTypes in the root Workspace.
func (m *Examples) GetWorkspaceTypes(ctx context.Context) (string, error) {
	return m.Kubectl(ctx).
		WithExec([]string{"kubectl", "get", "workspacetypes"}).
		Stdout(ctx)
}
