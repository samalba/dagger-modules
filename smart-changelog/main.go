// A generated module for Smartchangelog functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/smartchangelog/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type Smartchangelog struct{}

// Commit represents a Git commit with its metadata
type Commit struct {
	Hash        string       `json:"hash"`
	Subject     string       `json:"subject"`
	Body        string       `json:"body"`
	Author      string       `json:"author"`
	Date        string       `json:"date"`
	PullRequest *PullRequest `json:"pull_request,omitempty"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Description string `json:"body"`
	MergeCommit string `json:"merge_commit_sha"`
}

// ChangelogData represents all the information needed to generate a changelog
type ChangelogData struct {
	FromRef string   `json:"from_ref"`
	ToRef   string   `json:"to_ref"`
	Commits []Commit `json:"commits"`
}

// GetChangelogData gathers all commit and PR data between two refs
func (m *Smartchangelog) GetChangelogData(
	ctx context.Context,
	githubRepo string,
	fromRef string,
	toRef string,
) (*ChangelogData, error) {
	ghClient, err := NewGithubClient(githubRepo)
	if err != nil {
		return nil, err
	}

	commits, err := m.GetCommits(ctx, ghClient.GetGitURL(), fromRef, toRef)
	if err != nil {
		return nil, err
	}

	var enrichedCommits []Commit
	for _, commit := range commits {
		enriched, err := ghClient.EnrichCommitWithPR(ctx, commit)
		if err != nil {
			fmt.Printf("Warning: failed to fetch PR for commit %s: %v\n", commit.Hash, err)
			enrichedCommits = append(enrichedCommits, commit)
		} else {
			enrichedCommits = append(enrichedCommits, enriched)
		}
	}

	return &ChangelogData{
		FromRef: fromRef,
		ToRef:   toRef,
		Commits: enrichedCommits,
	}, nil
}

// GenerateChangelog generates a markdown changelog using Claude
func (m *Smartchangelog) GenerateChangelog(
	ctx context.Context,
	githubRepo string,
	fromRef string,
	toRef string,
	anthropicApiKey *dagger.Secret,
) (string, error) {
	data, err := m.GetChangelogData(ctx, githubRepo, fromRef, toRef)
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal changelog data: %w", err)
	}

	prompt := `Generate a well-formatted markdown changelog from this JSON data.
Group changes into sections: Features 🚀, Bug Fixes 🐛, Changes 🔄, and Maintenance 🧰.
Use commit subjects and PR information to create clear, user-focused descriptions.
Format the output as markdown with emojis and proper section headers.

JSON Data:
` + string(jsonData)

	apiKey, err := anthropicApiKey.Plaintext(ctx)
	if err != nil {
		return "", err
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.ModelClaude3_5SonnetLatest),
		MaxTokens: anthropic.F(int64(4096)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate changelog: %w", err)
	}

	return strings.TrimSpace(msg.Content[0].Text), nil
}
