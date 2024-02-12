package main

import (
	"context"
	"slices"
)

type InlinePython struct {
	Ctr *Container
}

func New(baseImage Optional[string]) *InlinePython {
	return &InlinePython{
		Ctr: dag.Container().From(baseImage.GetOr("python:3-alpine")),
	}
}

// fixme: add usage
func (m *InlinePython) WithPackage(name string) (*InlinePython, error) {
	m.Ctr = m.Ctr.WithExec([]string{"pip", "install", name})
	return m, nil
}

// fixme: add usage
func (m *InlinePython) WithPackages(packages []string) (*InlinePython, error) {
	// sort the requirements to optimize caching
	slices.Sort(packages)

	for _, name := range packages {
		m.WithPackage(name)
	}

	return m, nil
}

func (m *InlinePython) Code(ctx context.Context, code string) (*Container, error) {
	return m.Ctr.WithExec([]string{"python", "-c", code}).Sync(ctx)
}
