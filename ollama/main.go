// A generated module for Ollama functions
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
	"dagger/ollama/internal/dagger"
)

type Ollama struct{}

// Returns a container that echoes whatever string argument is provided
func (m *Ollama) ContainerEcho(stringArg string) *dagger.Container {
	return dag.Container().From("alpine:latest").WithExec([]string{"echo", stringArg})
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Ollama) DeployDaggerOnFly(ctx context.Context, token *dagger.Secret) (string, error) {
	dagrOnFly := dag.Dagrr(dagger.DagrrOpts{}).OnFlyio(token, dagger.DagrrOnFlyioOpts{
		Org: "dagger",
	})

	manifestDir := dagrOnFly.Manifest(dagger.DagrrFlyManifestOpts{
		Disk:          "32GB",
		Size:          "performance-2x",
		Memory:        "16GB",
		GpuKind:       "l40s",
		PrimaryRegion: "ord",
	})

	return dagrOnFly.Deploy(ctx, dagger.DagrrFlyDeployOpts{
		Dir: manifestDir,
	})
}
