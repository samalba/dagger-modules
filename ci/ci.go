package main

func (m *Ci) getBaseImage() *Container {
	ctr := dag.Container().From("alpine:3")

	ctr = ctr.WithMountedDirectory("/source", m.WorkDir)

	// Install Alpine dependencies
	ctr = ctr.WithExec([]string{
		"apk", "add", "--no-cache",
		"go",
		"curl",
	})

	// Install golangci-lint
	ctr = ctr.WithExec([]string{
		"sh", "-c",
		"curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.56.2",
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
