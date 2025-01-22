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
	Prefix      string       `json:"prefix"`
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
	URL         string `json:"url"`
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

func callClaude(ctx context.Context, anthropicApiKey *dagger.Secret, prompt string) (string, error) {
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
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	return strings.TrimSpace(msg.Content[0].Text), nil
}

func generateChangelogContent(ctx context.Context, anthropicApiKey *dagger.Secret, data *ChangelogData) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal changelog data: %w", err)
	}

	prompt := `Generate a well-formatted markdown changelog from this JSON data.
	Group changes into sections: Features 🚀, Bug Fixes 🐛, Changes 🔄, and Maintenance 🧰.
	Use commit subjects and PR information to create clear, user-focused descriptions.

	For each change, end the line with: "in {pr_url}", only if there's a pull request URL available.

	Format the output as markdown with emojis and proper section headers.

	JSON Data:
	` + string(jsonData)

	return callClaude(ctx, anthropicApiKey, prompt)
}

func generateBlogPost(ctx context.Context, anthropicApiKey *dagger.Secret, data *ChangelogData) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal changelog data: %w", err)
	}

	prompt := `Generate an engaging blog post announcing the changes in this release.
	Focus on the most important features and improvements.
	The tone should be professional but friendly.
	Include code examples or configuration snippets if relevant.
	Format the output as markdown.
	
	Structure the post with:
	1. A brief introduction
	2. Highlight of key features
	3. Notable improvements and fixes
	4. Conclusion with next steps or future plans

	Use the PR descriptions for detailed information when available.

	JSON Data:
	` + string(jsonData)

	return callClaude(ctx, anthropicApiKey, prompt)
}

// GenerateChangelog generates both a changelog and blog post using Claude
func (m *Smartchangelog) GenerateChangelog(
	ctx context.Context,
	githubRepo string,
	fromRef string,
	toRef string,
	anthropicApiKey *dagger.Secret,
) (*dagger.Directory, error) {
	data, err := m.GetChangelogData(ctx, githubRepo, fromRef, toRef)
	if err != nil {
		return nil, err
	}

	// Filter out commits with "chore" or "doc" prefixes
	filteredCommits := make([]Commit, 0, len(data.Commits))
	for _, commit := range data.Commits {
		if commit.Prefix != "chore" && commit.Prefix != "doc" {
			filteredCommits = append(filteredCommits, commit)
		}
	}
	data.Commits = filteredCommits

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal changelog data: %w", err)
	}

	changelog, err := generateChangelogContent(ctx, anthropicApiKey, data)
	if err != nil {
		return nil, err
	}

	blogpost, err := generateBlogPost(ctx, anthropicApiKey, data)
	if err != nil {
		return nil, err
	}

	return dag.Directory().
		WithNewFile("changelog.md", changelog).
		WithNewFile("blogpost.md", blogpost).
		WithNewFile("raw_data.json", string(jsonData)), nil
}
