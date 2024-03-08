package main

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
)

func (m *Ci) getBaseImage() *Container {
	ctr := dag.Container().From("alpine:3")

	ctr = ctr.WithMountedDirectory("/source", m.WorkDir)

	// Install Alpine dependencies
	ctr = ctr.WithExec([]string{
		"apk", "add", "--no-cache",
		"go",
		"curl",
	})

	// Env vars
	ctr = ctr.
		WithEnvVariable("GOPATH", "/go").
		WithEnvVariable("PATH", "/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")

	// Mount the code and set the work dir
	ctr = ctr.
		WithMountedDirectory("/source", m.WorkDir).
		WithWorkdir("/source")

	return ctr
}

func (m *Ci) GoLint(subdir string) *Container {
	if subdir == "" {
		subdir = "."
	}

	// Install golangci-lint
	ctr := m.getBaseImage().WithExec([]string{
		"sh", "-c",
		"curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.56.2",
	})

	return ctr.WithExec([]string{"sh", "-c", fmt.Sprintf("cd %q && golangci-lint run -v", subdir)})
}

func (m *Ci) PythonLint(subdir string) *Container {
	if subdir == "" {
		subdir = "."
	}

	// Install ruff from testing repository
	ctr := m.getBaseImage().WithExec([]string{
		"apk", "add", "--no-cache",
		"-X", "https://dl-cdn.alpinelinux.org/alpine/edge/testing/",
		"ruff",
	})

	return ctr.WithExec([]string{"ruff", "check", "-v", subdir})
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