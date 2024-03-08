package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v59/github"
)

func helpCommandsMessage() string {
	message := "\n\n---\n\n"
	message += "Available commands:\n\n"
	message += "- `!echo <message>`: Echoes a message\n"
	message += "- `!golint <subdir>`: Runs the Go linter on a sub directory\n"
	message += "- `!pythonlint <subdir>`: Runs the Python linter on a sub directory\n"
	message += "- `!event`: Shows the github event (debugging)\n"
	message += "- `!sh <command>`: Runs a shell command\n"

	return message
}

// checks if a user is authorized to run commands
func checkAuthorAssociation(ev *github.IssueCommentEvent) bool {
	authorized := []string{"OWNER", "MEMBER", "COLLABORATOR", "CONTRIBUTOR"}
	for _, a := range authorized {
		if ev.Comment.GetAuthorAssociation() == a {
			return true
		}
	}
	return false
}

func (m *Ci) handleIssueComment(ctx context.Context, githubToken *Secret, ev *github.IssueCommentEvent, eventData string) error {
	if !checkAuthorAssociation(ev) {
		return fmt.Errorf("User is not authorized to run commands")
	}

	command, args := "", ""
	parts := strings.SplitN(ev.Comment.GetBody(), " ", 2)
	if len(parts) >= 2 {
		args = parts[1]
	}
	command = parts[0]

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
		stdout, err := m.getBaseImage(m.WorkDir).WithExec([]string{"sh", "-c", args}).Stdout(ctx)
		if err != nil {
			_, err = comment.Create(ctx, fmt.Sprintf("`$ %s`\n\n```\n%s\n```", args, err.Error()))
			return err
		}
		_, err = comment.Create(ctx, fmt.Sprintf("`$ %s`\n\n```\n%s\n```", args, stdout))
		return err
	case "!event":
		if _, err := comment.Create(ctx, fmt.Sprintf("```json\n%s\n```", eventData)); err != nil {
			return err
		}
	case "!golint":
		if _, err := m.GoLint(args).Stdout(ctx); err != nil {
			_, err = comment.Create(ctx, fmt.Sprintf("Go linter failed: %s", err.Error()))
			return err
		}
		if _, err := comment.Create(ctx, "Go linter passed!"); err != nil {
			return err
		}
	case "!pythonlint":
		if _, err := m.PythonLint(args).Stdout(ctx); err != nil {
			_, err = comment.Create(ctx, fmt.Sprintf("Python linter failed: %s", err.Error()))
			return err
		}
		if _, err := comment.Create(ctx, "Python linter passed!"); err != nil {
			return err
		}
	}

	return nil
}
