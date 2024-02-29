package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/v59/github"
)

type Ci struct{}

func (m *Ci) Handle(ctx context.Context, githubToken *Secret, eventName string, eventFile *File) error {
	eventData, err := eventFile.Contents(ctx)
	if err != nil {
		return err
	}
	payload, err := github.ParseWebHook(eventName, []byte(eventData))
	if err != nil {
		return err
	}

	switch ev := payload.(type) {
	case *github.PullRequestEvent:
		switch ev.GetAction() {
		case "opened", "synchronize", "reopened", "ready_for_review":
			comment := dag.GithubComment(
				githubToken,
				ev.GetRepo().GetOwner().GetLogin(),
				ev.GetRepo().GetName(),
				GithubCommentOpts{
					Commit: ev.GetPullRequest().GetHead().GetSHA(),
				},
			)

			message := fmt.Sprintf("Hello @%s!\n\nThanks for opening a PR.", ev.PullRequest.User.GetLogin())
			if _, err := comment.Create(ctx, message); err != nil {
				return err
			}
			if _, err := comment.React(ctx, "rocket"); err != nil {
				return err
			}
		}
	}

	return nil
}
