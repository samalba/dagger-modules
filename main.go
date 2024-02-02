package main

import "context"

type Ci struct{}

func (m *Ci) RunTests() bool {
	ctx := context.Background()
	// FIXME: run tests in parallel + group errors
	m.TestInlinePython(ctx)
	return true
}

func (m *Ci) TestInlinePython(ctx context.Context) bool {
	res, err := dag.InlinePython().Code("print(True, end='')").Stdout(ctx)
	if err != nil {
		panic(err)
	}
	if res != "True" {
		panic(res)
	}
	return true
}
