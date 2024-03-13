package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v59/github"
)

type GithubComment struct {
	GithubToken *Secret
	MessageID   string
	Owner       string
	Repo        string
	Issue       int
	Commit      string
	CommentId   int64
}

func New(
	ctx context.Context,
	githubToken *Secret,
	// +optional
	// +default="github.com/aluzzardi/daggerverse/github-comment"
	messageID string,
	owner string,
	repo string,
	// +optional
	issue int,
	// +optional
	commit string,
	// +optional
	commentId int64,
) *GithubComment {
	return &GithubComment{
		GithubToken: githubToken,
		MessageID:   messageID,
		Owner:       owner,
		Repo:        repo,
		Issue:       issue,
		Commit:      commit,
		CommentId:   commentId,
	}
}

func (m *GithubComment) newClient(ctx context.Context) (*github.Client, error) {
	token, err := m.GithubToken.Plaintext(ctx)
	if err != nil {
		return nil, err
	}
	return github.NewClient(nil).WithAuthToken(token), nil
}

func marker(messageID string) string {
	return fmt.Sprintf("<!-- marker: %s -->", messageID)
}

func (m *GithubComment) markBody(body string) *string {
	marked := marker(m.MessageID) + "\n" + body
	return &marked
}

func (m *GithubComment) getIssueFromCommit(ctx context.Context, ghc *github.Client, commitSha string) (int, error) {
	prs, _, err := ghc.PullRequests.ListPullRequestsWithCommit(ctx, m.Owner, m.Repo, commitSha, nil)
	if err != nil {
		return 0, err
	}
	if len(prs) == 0 {
		return 0, fmt.Errorf("commit %s not found in any pull request", commitSha)
	}
	return prs[0].GetNumber(), nil
}

func (m *GithubComment) findComment(ctx context.Context, ghc *github.Client) (*github.IssueComment, int, error) {
	opt := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}
	var (
		issue = m.Issue
		err   error
	)
	if issue == 0 {
		if m.Commit == "" {
			return nil, 0, fmt.Errorf("either issue or commit must be set")
		}
		issue, err = m.getIssueFromCommit(ctx, ghc, m.Commit)
		if err != nil {
			return nil, 0, err
		}
	}
	for {
		comments, resp, err := ghc.Issues.ListComments(ctx, m.Owner, m.Repo, issue, opt)
		if err != nil {
			return nil, 0, err
		}

		for _, comment := range comments {
			if comment.Body != nil && strings.HasPrefix(*comment.Body, marker(m.MessageID)) {
				return comment, issue, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return nil, issue, nil
}

// example usage: "dagger call --github-token env:GITHUB_TOKEN --owner aluzzardi --repo daggerverse --issue 1 comment --body "hello world"
func (m *GithubComment) Create(ctx context.Context, body string) (*string, error) {
	ghc, err := m.newClient(ctx)
	if err != nil {
		return nil, err
	}
	existingComment, issue, err := m.findComment(ctx, ghc)
	if err != nil {
		return nil, err
	}

	var comment *github.IssueComment
	if existingComment != nil {
		existingComment.Body = m.markBody(body)
		comment, _, err = ghc.Issues.EditComment(ctx, m.Owner, m.Repo, *existingComment.ID, existingComment)
	} else {
		comment, _, err = ghc.Issues.CreateComment(ctx, m.Owner, m.Repo, issue, &github.IssueComment{
			Body: m.markBody(body),
		})
	}
	if err != nil {
		return nil, err
	}

	return comment.HTMLURL, nil
}

func (m *GithubComment) Append(ctx context.Context, body string) (*string, error) {
	ghc, err := m.newClient(ctx)
	if err != nil {
		return nil, err
	}

	comment, _, err := ghc.Issues.GetComment(ctx, m.Owner, m.Repo, m.CommentId)
	if err != nil {
		return nil, err
	}

	// Strip body from the marker onwards
	commentBody := strings.SplitN(*comment.Body, "\n"+marker(m.MessageID), 2)[0]

	body = fmt.Sprintf("\n\n---\n\n%s", body)
	newBody := fmt.Sprintf("%s\n%s", commentBody, *m.markBody(body))
	comment.Body = &newBody
	comment, _, err = ghc.Issues.EditComment(ctx, m.Owner, m.Repo, *comment.ID, comment)
	if err != nil {
		return nil, err
	}

	return comment.HTMLURL, nil
}

// example usage: "dagger call --github-token env:GITHUB_TOKEN --owner aluzzardi --repo daggerverse --issue 1 delete
func (m *GithubComment) Delete(ctx context.Context) error {
	ghc, err := m.newClient(ctx)
	if err != nil {
		return err
	}

	comment, _, err := m.findComment(ctx, ghc)
	if err != nil {
		return err
	}
	if comment == nil {
		return nil
	}

	_, err = ghc.Issues.DeleteComment(ctx, m.Owner, m.Repo, *comment.ID)
	return err
}

// example usage: "dagger call --github-token env:GITHUB_TOKEN --owner aluzzardi --repo daggerverse --issue 1 reaction +1
// The kind should be one of the following values: "+1", "-1", "laugh", "confused", "heart", "hooray", "rocket", or "eyes".
func (m *GithubComment) React(
	ctx context.Context,
	//	"+1", "-1", "laugh", "confused", "heart", "hooray", "rocket", or "eyes".
	kind string,
) error {
	ghc, err := m.newClient(ctx)
	if err != nil {
		return err
	}

	comment, _, err := m.findComment(ctx, ghc)
	if err != nil {
		return err
	}
	if comment == nil {
		return nil
	}

	_, _, err = ghc.Reactions.CreateIssueCommentReaction(ctx, m.Owner, m.Repo, *comment.ID, kind)
	return err
}
