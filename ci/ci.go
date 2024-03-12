package main

import (
	"context"
	"fmt"
	"main/internal/dagger"
	"strings"

	"github.com/acarl005/stripansi"
	"golang.org/x/sync/errgroup"
)

func (m *Ci) getBaseImage(sourceDir *dagger.Directory) *Container {
	ctr := dag.Container().From("alpine:3")

	ctr = ctr.WithMountedDirectory("/source", m.WorkDir)

	// Install Alpine dependencies
	ctr = ctr.WithExec([]string{
		"apk", "add", "--no-cache",
		"go",
		"curl",
		"docker",
	})

	// Env vars
	ctr = ctr.
		WithEnvVariable("GOPATH", "/go").
		WithEnvVariable("PATH", "/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")

	// Mount cache volumes
	ctr = ctr.
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("ci-go-pkg-mod")).
		WithMountedCache("/root/.cache", dag.CacheVolume("ci-root-cache"))

	// Mount the code and set the work dir
	ctr = ctr.
		WithMountedDirectory("/source", sourceDir).
		WithWorkdir("/source")

	return ctr
}

func (m *Ci) GoLint(subdir string) *Container {
	if subdir == "" {
		subdir = "."
	}

	// Install golangci-lint
	ctr := m.getBaseImage(m.WorkDir.Directory(subdir)).WithExec([]string{
		"sh", "-c",
		"curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.56.2",
	})

	return ctr.WithExec([]string{"golangci-lint", "run", "-v"})
}

func (m *Ci) PythonLint(subdir string) *Container {
	if subdir == "" {
		subdir = "."
	}

	// Install ruff from testing repository
	ctr := m.getBaseImage(m.WorkDir.Directory(subdir)).WithExec([]string{
		"apk", "add", "--no-cache",
		"-X", "https://dl-cdn.alpinelinux.org/alpine/edge/testing/",
		"ruff",
	})

	return ctr.WithExec([]string{"ruff", "check", "-v"})
}

func (m *Ci) RunAllLinters(ctx context.Context) (string, error) {
	goModules := []string{"ci", "inline-python", "remote-dagger", "ssh"}
	pythonModules := []string{"hello-world"}

	g, ctx := errgroup.WithContext(ctx)

	for _, mod := range goModules {
		modName := mod
		g.Go(func() error {
			if _, err := m.GoLint(modName).Stdout(ctx); err != nil {
				return err
			}
			return nil
		})
	}

	for _, mod := range pythonModules {
		modName := mod
		g.Go(func() error {
			if _, err := m.PythonLint(modName).Stdout(ctx); err != nil {
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return "", err
	}

	return fmt.Sprintf("All linters passed: %s", strings.Join(append(goModules, pythonModules...), ", ")), nil
}

func (m *Ci) DaggerCLI(ctx context.Context, args string) (string, error) {
	ctr := m.getBaseImage(m.WorkDir).
		WithExec([]string{"sh", "-c", "curl -L https://dl.dagger.io/dagger/install.sh | BIN_DIR=/bin sh"}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("2>&1 dagger --silent %s ; true", args)}, dagger.ContainerWithExecOpts{
			ExperimentalPrivilegedNesting: true,
		})
	out, err := ctr.Stdout(ctx)
	if err != nil {
		return "", err
	}
	// Strip ANSI escape codes
	out = stripansi.Strip(out)
	return out, nil
}
