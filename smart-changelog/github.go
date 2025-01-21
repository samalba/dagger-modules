package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GithubClient handles GitHub API operations
type GithubClient struct {
	owner string
	repo  string
}

// NewGithubClient creates a new GitHub client from a repository string (github.com/owner/repo)
func NewGithubClient(githubRepo string) (*GithubClient, error) {
	parts := strings.Split(githubRepo, "/")
	if len(parts) != 3 || parts[0] != "github.com" {
		return nil, fmt.Errorf("invalid GitHub repository format. Expected 'github.com/org/repo', got '%s'", githubRepo)
	}

	return &GithubClient{
		owner: parts[1],
		repo:  parts[2],
	}, nil
}

// GetGitURL returns the Git clone URL for the repository
func (g *GithubClient) GetGitURL() string {
	return fmt.Sprintf("https://github.com/%s/%s.git", g.owner, g.repo)
}

// GetPullRequestForCommit fetches the associated pull request for a commit
func (g *GithubClient) GetPullRequestForCommit(ctx context.Context, commitHash string) (*PullRequest, error) {
	output, err := dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		WithExec([]string{"apk", "add", "--no-cache", "curl", "jq"}).
		WithEnvVariable("COMMIT_HASH", commitHash).
		WithEnvVariable("OWNER", g.owner).
		WithEnvVariable("REPO", g.repo).
		WithExec([]string{"sh", "-c", `
			curl -s "https://api.github.com/repos/$OWNER/$REPO/commits/$COMMIT_HASH/pulls" \
				-H "Accept: application/vnd.github.v3+json" |
			jq -r 'if length > 0 then .[0] | {number, title, body, merge_commit_sha} else empty end'
		`}).
		Stdout(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch pull request: %w", err)
	}

	if output == "" {
		return nil, nil // No pull request found for this commit
	}

	var pr PullRequest
	if err := json.Unmarshal([]byte(output), &pr); err != nil {
		return nil, fmt.Errorf("failed to parse pull request data: %w", err)
	}

	return &pr, nil
}

// EnrichCommitWithPR fetches PR information for a commit and returns a new enriched commit
func (g *GithubClient) EnrichCommitWithPR(ctx context.Context, commit Commit) (Commit, error) {
	pr, err := g.GetPullRequestForCommit(ctx, commit.Hash)
	if err != nil {
		return commit, err
	}

	enriched := commit
	enriched.PullRequest = pr
	return enriched, nil
}

// // GenerateChangeMessage uses Claude to create a user-friendly changelog message
// func (g *GithubClient) GenerateChangeMessage(ctx context.Context, anthropicApiKey *dagger.Secret, commit Commit) (string, error) {
// 	var prompt strings.Builder
// 	prompt.WriteString("Generate a clear, concise changelog entry based on this information.\n")
// 	prompt.WriteString("Focus on explaining the change from a user's perspective.\n\n")

// 	prompt.WriteString("Commit message:\n")
// 	prompt.WriteString(commit.Subject + "\n")
// 	if commit.Body != "" {
// 		prompt.WriteString(commit.Body + "\n")
// 	}

// 	if commit.PullRequest != nil {
// 		prompt.WriteString("\nPull Request #" + fmt.Sprint(commit.PullRequest.Number) + ":\n")
// 		prompt.WriteString("Title: " + commit.PullRequest.Title + "\n")
// 		if commit.PullRequest.Description != "" {
// 			prompt.WriteString("Description: " + commit.PullRequest.Description + "\n")
// 		}
// 	}

// 	apiKey, err := anthropicApiKey.Plaintext(ctx)
// 	if err != nil {
// 		return "", err
// 	}

// 	client := anthropic.NewClient(
// 		option.WithAPIKey(apiKey),
// 	)

// 	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
// 		Model:     anthropic.F(anthropic.ModelClaude3_5SonnetLatest),
// 		MaxTokens: anthropic.F(int64(1024)),
// 		Messages: anthropic.F([]anthropic.MessageParam{
// 			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt.String())),
// 		}),
// 	})

// 	if err != nil {
// 		return "", fmt.Errorf("failed to generate change message: %w", err)
// 	}

// 	return strings.TrimSpace(msg.Content[0].Text), nil
// }
