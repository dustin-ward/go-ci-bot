package handlers

import (
	"log"
	"test-org-gozbot/build"
	"test-org-gozbot/gh/auth"

	"github.com/google/go-github/v62/github"
)

func HandlePullRequestEvent(event *github.PullRequestEvent) error {
	//TODO: Remove
	apiClient, err := auth.GetClient()
	if err != nil {
		return err
	}

	action := event.GetAction()
	if action == "opened" || action == "reopened" {
		pr := event.GetPullRequest()
		log.Printf("Pull request opened: #%d %s - %s (%s)\n",
			pr.GetNumber(),
			pr.GetTitle(),
			pr.GetUser().GetLogin(),
			pr.GetHead().GetSHA()[:6],
		)

		ok, err := build.Push(apiClient, pr.GetNumber(), pr.GetHead().GetSHA(), pr.GetHead().GetUser().GetLogin())
		if err != nil {
			return err
		}
		if ok {
			log.Println("Added to build queue")
		} else {
			log.Println("Not added to build queue")
		}
	}

	return nil
}
