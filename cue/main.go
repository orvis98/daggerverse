// A generated module for Cue functions
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
	"dagger/cue/internal/dagger"
	"fmt"
	"strings"
)

type Cue struct {
	// Container with the cue binary
	Container *dagger.Container
}

func New(
	// +optional
	// +default="latest"
	// the desired CUE version
	version string,
) *Cue {
	return &Cue{
		Container: dag.Container().
			From("golang").
			WithExec([]string{"go", "install", fmt.Sprintf("cuelang.org/go/cmd/cue@%s", version)}),
	}
}

// Execute CUE
func (c *Cue) Exec(
	ctx context.Context,
	// the args to CUE
	command string,
	// the CUE context directory
	context *dagger.Directory,
) *dagger.Container {
	return c.Container.
		WithDirectory("cue", context).
		WithWorkdir("cue").
		WithExec(append([]string{"cue"}, strings.Split(command, " ")...))
}
