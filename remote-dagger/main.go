package main

import (
	"context"
	"fmt"
	"strings"
)

type RemoteDagger struct {
	SSHClient *SSHClient
}

func New(destination string, identityFile *Secret) (*RemoteDagger, error) {
	return &RemoteDagger{
		SSHClient: NewSSHClient(destination, identityFile),
	}, nil
}

func (m *RemoteDagger) run(args []string) *Container {
	return m.SSHClient.Command(args)
}

// example usage: "dagger call --destination USER@HOST --identity-file file:${HOME}/.ssh/id_ed25519 install"
func (m *RemoteDagger) SetupDagger(ctx context.Context) error {
	runArgs := []string{"/bin/bash", "-c", "\"~/.local/bin/dagger version || curl -L https://dl.dagger.io/dagger/install.sh | BIN_DIR=$HOME/.local/bin sh && ~/.local/bin/dagger version\""}
	stderr, err := m.run(runArgs).Stderr(ctx)
	if err != nil {
		return fmt.Errorf("failed to install dagger: %v - %s", err, stderr)
	}
	return nil
}

// example usage: "dagger call --destination USER@HOST --identity-file file:${HOME}/.ssh/id_ed25519 remote-call ... stdout"
func (m *RemoteDagger) RemoteCall(ctx context.Context, module string, args []string) (*Container, error) {
	if err := m.SetupDagger(ctx); err != nil {
		return nil, err
	}

	runArgs := []string{"/bin/bash", "-c", fmt.Sprintf("\"~/.local/bin/dagger --mod '%s' call %s\"", module, strings.Join(args, " "))}
	return m.SSHClient.Command(runArgs), nil
}
