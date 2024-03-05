package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v59/github"
)

func (m *Ci) handleIssueComment(ctx context.Context, githubToken *Secret, ev *github.IssueCommentEvent) error {
	// FIXME: check if author is a contributor

	parts := strings.SplitN(ev.Comment.GetBody(), " ", 2)
	if len(parts) != 2 {
		return nil
	}
	command, args := parts[0], parts[1]

	comment := dag.GithubComment(
		githubToken,
		ev.GetRepo().GetOwner().GetLogin(),
		ev.GetRepo().GetName(),
		GithubCommentOpts{
			Issue: ev.Issue.GetNumber(),
		},
	)

	switch command {
	case "!echo":
		if _, err := comment.Create(ctx, args); err != nil {
			return err
		}
	case "!sh":
		stdout, err := m.getBaseImage().Stdout(ctx)
		if err != nil {
			_, err = comment.Create(ctx, fmt.Sprintf("`%s`: %s", args, err.Error()))
			return err
		}
		_, err = comment.Create(ctx, fmt.Sprintf("`$ %s`\n\n```%s```", args, stdout))
		return err
	}

	return nil
}
