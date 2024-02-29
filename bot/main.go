package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v59/github"
)

type GithubCi struct{}

func (m *GithubCi) Handle(ctx context.Context, githubToken *Secret, eventName string, eventFile *File) error {
	eventData, err := eventFile.Contents(ctx)
	if err != nil {
		return err
	}
	payload, err := github.ParseWebHook(eventName, []byte(eventData))
	if err != nil {
		return err
	}

	switch ev := payload.(type) {
	case *github.IssueCommentEvent:
		switch ev.GetAction() {
		case "created":
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
				stdout, err := dag.
					Container().
					From("alpine").
					WithExec([]string{"sh", "-c", args}).
					Stdout(ctx)
				if err != nil {
					_, err = comment.Create(ctx, fmt.Sprintf("`%s`: %s", args, err.Error()))
					return err
				}
				_, err = comment.Create(ctx, fmt.Sprintf("`$ %s`\n\n```%s```", args, stdout))
			}
		}
	}

	return nil
}
