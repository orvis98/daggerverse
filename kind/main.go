// A generated module for Kind functions
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
	"context"
	"dagger/kind/internal/dagger"
	"fmt"
	"strings"
)

type Kind struct {
	// Container with the kind and docker binaries
	Container *dagger.Container
}

func New(
	// +optional
	// +default="latest"
	// the desired kind version
	version string,
	// path to the host Docker socket
	dockerSocket *dagger.Socket,
) *Kind {
	docker := dag.Container().
		From("docker:dind").
		File("/usr/local/bin/docker")
	return &Kind{
		Container: dag.Container().
			From("golang").
			WithFile("/bin/docker", docker).
			WithExec([]string{"go", "install", "sigs.k8s.io/kind@" + version}).
			WithUnixSocket("/var/run/docker.sock", dockerSocket),
	}
}

// Prints the kind CLI version
func (k *Kind) Version(ctx context.Context) (string, error) {
	return k.Container.WithExec([]string{"kind", "version"}).
		Stdout(ctx)
}

// Lists existing kind clusters by their name
func (k *Kind) GetClusters(ctx context.Context) (string, error) {
	return k.Container.WithExec([]string{"kind", "get", "clusters"}).
		Stdout(ctx)
}

// Creates a local Kubernetes cluster
func (k *Kind) CreateCluster(
	ctx context.Context,
	// +optional
	// +default="kind"
	// cluster name, overrides KIND_CLUSTER_NAME, config
	name string,
	// +optional
	// path to a kind config file
	config *dagger.File,
	// +optional
	// node docker image to use for booting the cluster
	image string,
) (string, error) {
	clusters, err := k.GetClusters(ctx)
	if err != nil {
		return "", err
	}
	for _, c := range strings.Split(clusters, "\n") {
		if c == name {
			return "", fmt.Errorf("Cluster with name '%s' already exists", name)
		}
	}
	var flags []string
	ctr := k.Container
	if config != nil {
		ctr = ctr.WithFile("cluster.yaml", config)
		flags = append(flags, "--config", "cluster.yaml")
	}
	if image != "" {
		flags = append(flags, "--image", image)
	}
	return ctr.WithExec(append([]string{"kind", "create", "cluster", "--name", name}, flags...)).
		Stdout(ctx)
}

// Deletes a cluster
func (k *Kind) DeleteCluster(
	ctx context.Context,
	// +optional
	// +default="kind"
	// the cluster name
	name string,
) (string, error) {
	clusters, err := k.GetClusters(ctx)
	if err != nil {
		return "", err
	}
	for _, c := range strings.Split(clusters, "\n") {
		if c == name {
			return "", fmt.Errorf("Cluster with name '%s' does not exists", name)
		}
	}
	return k.Container.WithExec([]string{"kind", "delete", "cluster", "--name", name}).
		Stdout(ctx)
}

// Exports cluster kubeconfig
func (k *Kind) Kubeconfig(
	ctx context.Context,
	// +optional
	// +default="kind"
	// the cluster name
	name string,
) (*dagger.File, error) {
	clusters, err := k.GetClusters(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range strings.Split(clusters, "\n") {
		if c == name {
			return nil, fmt.Errorf("Cluster with name '%s' does not exists", name)
		}
	}
	return k.Container.WithExec([]string{"kind", "export", "kubeconfig", "--kubeconfig", "config", "--name", name}).
		File("config"), nil
}

// Exports logs to a directory
func (k *Kind) Logs(
	ctx context.Context,
	// +optional
	// +default="kind"
	// the cluster name
	name string,
) (*dagger.Directory, error) {
	clusters, err := k.GetClusters(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range strings.Split(clusters, "\n") {
		if c == name {
			return nil, fmt.Errorf("Cluster with name '%s' does not exists", name)
		}
	}
	return k.Container.WithExec([]string{"kind", "export", "logs", "logs", "--name", name}).
		Directory("logs"), nil
}
