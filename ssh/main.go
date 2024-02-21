// Run a command on a remote machine using SSH and return the result
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type SSH struct {
	BaseCtr     *Container
	Destination string
	Opts        []SSHOpts
	Cache       bool
}

type SSHOpts struct {
	IdentityFile *Secret
	Port         int
	Login        string
}

// FIXME: custom struct as args is currently unsupported
// (Error: unsupported list of objects "SshOpts" for flag: opts)
// Keeping the code for later
//
// // example usage: "dagger call --destination machine.localdomain"
// func New(destination string, opts ...SshOpts) (*Ssh, error) {
// 	baseCtr := dag.Container().From("alpine:3").WithExec([]string{"apk", "add", "--no-cache", "openssh-client"})
// 	ssh := &Ssh{
// 		Destination: destination,
// 		Opts:        opts,
// 		BaseCtr:     baseCtr,
// 	}
// 	return ssh, nil
// }

func New(
	// Destination to connect to (SSH destination)
	destination string,
	// Private key to connect
	identityFile *Secret,
	// Enable caching of commands
	// +optional
	// +default=false
	cache bool) (*SSH, error) {
	baseCtr := dag.Container().From("alpine:3").WithExec([]string{"apk", "add", "--no-cache", "openssh-client"})

	// FIXME: Currently only supporting few commands, see comments and FIXME above
	opts := SSHOpts{
		IdentityFile: identityFile,
	}

	ssh := &SSH{
		Destination: destination,
		Opts:        []SSHOpts{opts},
		BaseCtr:     baseCtr,
		Cache:       cache,
	}
	return ssh, nil
}

// example usage: "dagger call --destination USER@HOST --identity-file file:${HOME}/.ssh/id_ed25519 command --args whoami stdout"
func (m *SSH) makeCtrArgs() (*Container, []string) {
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

	// disables the cache by default (given we're calling remote commands)
	if m.Cache == false {
		ctr = ctr.WithEnvVariable("_SSH_CACHE_BUSTER", fmt.Sprintf("%d-%d", os.Getpid(), time.Now().UnixNano()))
	}

	// add the destination address after the ssh args
	execArgs = append(execArgs, m.Destination)

	return ctr, execArgs
}

// example usage: "dagger call --destination USER@HOST --identity-file file:${HOME}/.ssh/id_ed25519 command --args whoami stdout"
func (m *SSH) Command(args []string) *Container {
	ctr, execArgs := m.makeCtrArgs()
	// add the command args
	execArgs = append(execArgs, args...)

	return ctr.WithExec(execArgs)
}

// execute a remote script. similar to `command', but the script is piped to ssh (shell escaping safe)
func (m *SSH) Script(content string) *Container {
	ctr, execArgs := m.makeCtrArgs()

	return ctr.WithExec(execArgs, ContainerWithExecOpts{Stdin: content})
}
