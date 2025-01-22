package main

import (
	"context"
	"fmt"
	"strings"
)

// getCommitPrefix extracts and normalizes the commit type prefix
func getCommitPrefix(subject string) string {
	subject = strings.ToLower(subject)
	subject = strings.TrimSpace(subject)

	// Find the colon separator, accounting for optional spaces
	idx := strings.Index(subject, ":")
	if idx < 0 {
		return ""
	}

	// Extract the prefix part
	prefix := strings.TrimSpace(subject[:idx])

	// Handle scoped prefixes like "chore(deps)"
	if parenIdx := strings.Index(prefix, "("); parenIdx > 0 {
		prefix = strings.TrimSpace(prefix[:parenIdx])
	}

	switch prefix {
	case "feat", "feature":
		return "feat"
	case "fix", "bug":
		return "fix"
	case "doc", "docs":
		return "doc"
	case "chore", "build", "ci":
		return "chore"
	case "test", "tests":
		return "test"
	case "refactor":
		return "refactor"
	case "style":
		return "style"
	case "perf":
		return "perf"
	}
	return ""
}

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
			Prefix:  getCommitPrefix(lines[1]),
		}
		commits = append(commits, commit)
	}

	return commits, nil
}
