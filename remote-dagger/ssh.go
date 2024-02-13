package main

import (
	"fmt"
	"strconv"
)

type SSHClient struct {
	BaseCtr     *Container
	Destination string
	Opts        []SSHOpts
}

type SSHOpts struct {
	IdentityFile *Secret
	Port         int
	Login        string
}

func NewSSHClient(destination string, identityFile *Secret) *SSHClient {
	baseCtr := dag.Container().From("alpine:3").WithExec([]string{"apk", "add", "--no-cache", "openssh-client"})

	// FIXME: Currently only supporting few commands, see comments and FIXME above
	opts := SSHOpts{
		IdentityFile: identityFile,
	}

	return &SSHClient{
		Destination: destination,
		Opts:        []SSHOpts{opts},
		BaseCtr:     baseCtr,
	}
}

// example usage: "dagger call --destination USER@HOST --identity-file file:${HOME}/.ssh/id_ed25519 command --args whoami stdout"
func (m *SSHClient) Command(args []string) *Container {
	ctr := m.BaseCtr

	execArgs := []string{"/usr/bin/ssh", "-o", "StrictHostKeyChecking=no"}
	for i, o := range m.Opts {
		if o.IdentityFile != nil {
			// this allows to support several keys if many opts were passed
			keyPath := fmt.Sprintf("/key_%d", i)
			ctr = ctr.WithMountedSecret(keyPath, o.IdentityFile)
			execArgs = append(execArgs, "-i", keyPath)
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
