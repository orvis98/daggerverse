// A generated module for Oscal functions
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
	"dagger/oscal/internal/dagger"
	"fmt"
)

type Oscal struct {
	// returns the cue version
	CueVersion string
	// +private
	Source *dagger.Directory
}

func New(
	// +optional
	// +default="v0.12.0"
	// the desired CUE version
	cueVersion string,
	// +optional
	// +defaultPath="/oscal"
	// the context directory
	source *dagger.Directory,
) *Oscal {
	return &Oscal{
		CueVersion: cueVersion,
		Source:     source,
	}
}

// returns a cue container
func (m *Oscal) Container() *dagger.Container {
	return dag.Container().
		From("golang").
		WithExec([]string{"go", "install", fmt.Sprintf("cuelang.org/go/cmd/cue@%s", m.CueVersion)}).
		WithDirectory("/src", m.Source).
		WithWorkdir("/src")
}

// returns the component definition as a string
func (m *Oscal) ComponentDefinition(
	ctx context.Context,
	// the component definition
	file *dagger.File,
) (string, error) {
	return m.Container().
		WithFile("component.cue", file).
		WithExec([]string{"cue", "cmd", "export"}).
		Stdout(ctx)
}
