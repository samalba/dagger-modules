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
	"fmt"
)

type Gpu struct{}

// Returns lines that match a pattern in the files of the provided Directory
// Once deployed, in order to point to the new remote engine:
// export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://<FLY_APP_NAME>.internal:2345
func (m *Gpu) DeployDaggerOnFly(ctx context.Context, token *dagger.Secret, org string) (string, error) {
	dagr := dag.Dagrr(dagger.DagrrOpts{})
	dagrOnFly := dagr.OnFlyio(token, dagger.DagrrOnFlyioOpts{
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

	if _, err := dagrOnFly.Deploy(ctx, dagger.DagrrFlyDeployOpts{
		Dir: manifestDir,
	}); err != nil {
		return "", err
	}

	appName, _ := dagr.GetApp(ctx)
	return fmt.Sprintf("export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://%s.internal:2345", appName), nil
}

// TestCuda tests if it can access the GPU, requires a machine with an NVIDIA GPU
func (m *Gpu) TestCuda(ctx context.Context) (string, error) {
	return dag.Container().
		From("nvidia/cuda:12.6.2-base-ubuntu24.04").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "libnvidia-compute-565"}).
		ExperimentalWithAllGPUs().
		WithExec([]string{"sh", "-c", "nvidia-smi -L"}).
		Stdout(ctx)
}

func (m *Gpu) OllamaWithGPU(ctx context.Context) (string, error) {
	ctr := dag.Container().
		From("nvidia/cuda:12.6.2-base-ubuntu24.04").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "libnvidia-compute-565"}).
		ExperimentalWithAllGPUs().
		WithMountedFile("/tmp/ollama.tgz", dag.HTTP("https://ollama.com/download/ollama-linux-amd64.tgz")).
		WithExec([]string{"tar", "-C", "/usr", "-xzf", "/tmp/ollama.tgz"})

	cache := dag.CacheVolume("ollama-data")

	server := ctr.
		WithMountedCache("/root/.ollama", cache).
		WithEnvVariable("OLLAMA_HOST", "0.0.0.0:11434").
		WithExec([]string{"ollama", "serve"}).
		WithExposedPort(11434).AsService()

	return ctr.
		WithServiceBinding("ollama", server).
		WithEnvVariable("OLLAMA_HOST", "ollama:11434").
		WithExec([]string{"ollama", "pull", "llama3.2"}).
		WithExec([]string{"ollama", "run", "llama3.2", "What color is the grass?"}).Stdout(ctx)
}

func (m *Gpu) OllamaWithCPU(ctx context.Context) (string, error) {
	ctr := dag.Container().From("ollama/ollama:0.4.7")

	cache := dag.CacheVolume("ollama-data")

	server := ctr.
		WithMountedCache("/root/.ollama", cache).
		WithExec([]string{"ollama", "serve"}).
		WithExposedPort(11434).AsService()

	return ctr.
		WithServiceBinding("ollama", server).
		WithEnvVariable("OLLAMA_HOST", "ollama:11434").
		WithExec([]string{"ollama", "pull", "llama3.2"}).
		WithExec([]string{"ollama", "run", "llama3.2", "What color is the grass?"}).Stdout(ctx)
}

// Destroy the remote Flyio app
func (m *Gpu) DestroyDaggerOnFly(ctx context.Context, token *dagger.Secret, app string) (string, error) {
	return dag.Flyio(token).Destroy(ctx, app)
}
