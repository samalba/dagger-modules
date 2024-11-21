// A generated module for Gpu functions
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
	"dagger/gpu/internal/dagger"
)

type Gpu struct{}

// Returns lines that match a pattern in the files of the provided Directory
// Once deployed, in order to point to the new remote engine:
// export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://<FLY_APP_NAME>.internal:2345
func (m *Gpu) DeployDaggerOnFly(ctx context.Context, token *dagger.Secret) (string, error) {
	dagrOnFly := dag.Dagrr(dagger.DagrrOpts{}).OnFlyio(token, dagger.DagrrOnFlyioOpts{
		Org: "dagger",
	})

	manifestDir := dagrOnFly.Manifest(dagger.DagrrFlyManifestOpts{
		Disk:          "32GB",
		Size:          "performance-2x",
		Memory:        "16GB",
		GpuKind:       "l40s",
		PrimaryRegion: "ord",
		Environment:   []string{"_EXPERIMENTAL_DAGGER_GPU_SUPPORT = \"true\""},
	})

	return dagrOnFly.Deploy(ctx, dagger.DagrrFlyDeployOpts{
		Dir: manifestDir,
	})
}

// TestCuda tests if it can access the GPU, requires a machine with an NVIDIA GPU
func (m *Gpu) TestCuda(ctx context.Context) (string, error) {
	return dag.Container().
		// From("nvidia/cuda:12.6.2-base-ubuntu24.04").
		From("nvidia/cuda:11.7.1-base-ubuntu20.04").
		// FIXME: this is an attempt to fix the error: "fork/exec /usr/bin/nvidia-container-runtime-hook: no such file or directory"
		WithExec([]string{"sh", "-c", "apt update && apt install -y nvidia-container-toolkit"}).
		ExperimentalWithAllGPUs().
		WithExec([]string{"nvidia-smi", "-L"}).
		Stdout(ctx)
}
