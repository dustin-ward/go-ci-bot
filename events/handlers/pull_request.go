package handlers

import (
	"github.ibm.com/open-z/jeff-ci/tasks"
	"log"

	"github.com/google/go-github/v62/github"
)

func handlePullRequestEvent(event *github.PullRequestEvent) error {
	action := event.GetAction()
	if action == "opened" || action == "reopened" {
		return registerNewPR(event)
	}

	return nil
}

func registerNewPR(event *github.PullRequestEvent) error {
	pr := event.GetPullRequest()
	log.Printf("Pull request opened: #%d %s - %s (%s)\n",
		pr.GetNumber(),
		pr.GetTitle(),
		pr.GetUser().GetLogin(),
		pr.GetHead().GetSHA()[:6],
	)

	//TODO: Redundant code here and in pull_request.go?
	tasks.Build{
		PR:          pr.GetNumber(),
		Branch:      pr.GetHead().GetRef(),
		SHA:         pr.GetHead().GetSHA(),
		SubmittedBy: pr.GetHead().GetUser().GetLogin(),
	}.Enqueue()

	return nil
}
