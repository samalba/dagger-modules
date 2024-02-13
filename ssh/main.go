package main

import (
	"fmt"
	"strconv"
)

type Ssh struct {
	BaseCtr     *Container
	Destination string
	Opts        []SshOpts
}

type SshOpts struct {
	IdentityFile *Secret
	Port         int
	Login        string
}

// example usage: "dagger call --destination machine.localdomain"
func New(destination string, opts ...SshOpts) (*Ssh, error) {
	baseCtr := dag.Container().From("alpine:3").WithExec([]string{"apk", "add", "--no-cache", "openssh-client"})
	ssh := &Ssh{
		Destination: destination,
		Opts:        opts,
		BaseCtr:     baseCtr,
	}
	return ssh, nil
}

// example usage: "dagger call --destination machine.localdomain command whoami
func (m *Ssh) Command(args ...string) *Container {
	ctr := m.BaseCtr

	execArgs := []string{"/usr/bin/ssh"}
	for i, o := range m.Opts {
		if o.IdentityFile != nil {
			// this allows to support several keys if many opts were passed
			keyPath := fmt.Sprintf("/key_{%d}", i)
			ctr.WithMountedSecret(keyPath, o.IdentityFile)
			args = append(execArgs, "-i", keyPath)
		}
		if o.Login != "" {
			execArgs = append(execArgs, "-l", o.Login)
		}
		if o.Port > 0 {
			execArgs = append(execArgs, "-p", strconv.Itoa(o.Port))
		}
	}

	// add the destination address after the ssh args
	execArgs = append(execArgs, m.Destination)
	// add the command args
	execArgs = append(execArgs, args...)

	return ctr.WithExec(execArgs)
}
