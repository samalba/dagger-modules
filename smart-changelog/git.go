package main

import (
	"context"
	"fmt"
	"strings"
)

// GetCommits retrieves commits between two refs from a Git repository
func (m *Smartchangelog) GetCommits(
	ctx context.Context,
	repository string,
	fromRef string,
	toRef string,
) ([]Commit, error) {
	output, err := dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		// Install required tools
		WithExec([]string{"apk", "add", "--no-cache", "git"}).
		// Clone the repository
		WithExec([]string{"git", "clone", repository, "/repo"}).
		WithWorkdir("/repo").
		// Get commit logs between refs
		WithExec([]string{"git", "log", "--pretty=format:%H%n%s%n%b%n%aN%n%aI%n---COMMIT---", fromRef + ".." + toRef}).
		Stdout(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	// Split output into individual commits
	commitStrings := strings.Split(output, "---COMMIT---")
	var commits []Commit

	// Process each commit
	for _, commitStr := range commitStrings {
		if strings.TrimSpace(commitStr) == "" {
			continue
		}

		// Split commit info into lines
		lines := strings.Split(strings.TrimSpace(commitStr), "\n")
		if len(lines) < 5 {
			continue // Skip malformed commits
		}

		commit := Commit{
			Hash:    lines[0],
			Subject: lines[1],
			Body:    strings.TrimSpace(strings.Join(lines[2:len(lines)-2], "\n")),
			Author:  lines[len(lines)-2],
			Date:    lines[len(lines)-1],
		}
		commits = append(commits, commit)
	}

	return commits, nil
}
