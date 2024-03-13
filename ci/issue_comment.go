package main

import (
	"context"
	"fmt"
	"main/internal/dagger"
	"strings"

	"github.com/google/go-github/v59/github"
)

func helpCommandsMessage() string {
	message := "\n\n---\n\n"
	message += "Available commands:\n\n"
	message += "- `!echo <message>`: Echoes a message\n"
	message += "- `!golint <subdir>`: Runs the Go linter on a sub directory\n"
	message += "- `!pythonlint <subdir>`: Runs the Python linter on a sub directory\n"
	message += "- `!dagger <command>`: Runs a dagger command\n"
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

func parseCommandArgs(body string) (string, string) {
	command, args := "", ""

	// only keep the first line of the comment
	parts := strings.SplitN(body, "\n", 2)
	if len(parts) < 1 {
		return command, args
	}
	body = parts[0]

	body = strings.TrimSpace(body)
	parts = strings.SplitN(body, " ", 2)
	if len(parts) >= 2 {
		args = parts[1]
	}
	command = parts[0]

	args = strings.TrimSpace(args)
	command = strings.TrimSpace(command)

	return command, args
}

func updateComment(ctx context.Context, comment *dagger.GithubComment, message string) error {
	if _, err := comment.Append(ctx, message); err != nil {
		return err
	}

	return nil
}

func (m *Ci) handleIssueComment(ctx context.Context, githubToken *Secret, ev *github.IssueCommentEvent, eventData string) error {
	if !checkAuthorAssociation(ev) {
		return fmt.Errorf("User is not authorized to run commands")
	}

	command, args := parseCommandArgs(ev.Comment.GetBody())

	comment := dag.GithubComment(
		githubToken,
		ev.GetRepo().GetOwner().GetLogin(),
		ev.GetRepo().GetName(),
		GithubCommentOpts{
			Issue:     ev.Issue.GetNumber(),
			CommentID: int(ev.Comment.GetID()),
		},
	)

	switch command {
	case "!echo":
		return updateComment(ctx, comment, args)
	case "!sh":
		stdout, err := m.getBaseImage(m.WorkDir).WithExec([]string{"sh", "-c", args}).Stdout(ctx)
		message := fmt.Sprintf("`$ %s`\n\n```\n%s\n```", args, stdout)
		if err != nil {
			message = fmt.Sprintf("`$ %s`\n\n```\n%s\n```", args, err.Error())
		}
		return updateComment(ctx, comment, message)
	case "!event":
		return updateComment(ctx, comment, fmt.Sprintf("```json\n%s\n```", eventData))
	case "!golint":
		if _, err := m.GoLint(args).Stdout(ctx); err != nil {
			return updateComment(ctx, comment, fmt.Sprintf("Go linter failed: %s", err.Error()))
		}
		return updateComment(ctx, comment, "Go linter passed!")
	case "!pythonlint":
		if _, err := m.PythonLint(args).Stdout(ctx); err != nil {
			return updateComment(ctx, comment, fmt.Sprintf("Python linter failed: %s", err.Error()))
		}
		return updateComment(ctx, comment, "Python linter passed!")
	case "!dagger":
		stdout, err := m.DaggerCLI(ctx, args)
		message := fmt.Sprintf("`$ dagger %s`\n\n```\n%s\n```", args, stdout)
		if err != nil {
			message = fmt.Sprintf("`$ dagger %s`\n\n```\n%s\n```", args, err.Error())
		}
		return updateComment(ctx, comment, message)
	}

	return nil
}
