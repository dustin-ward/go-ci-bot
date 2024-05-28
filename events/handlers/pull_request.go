package handlers

import (
	"log"
	"test-org-gozbot/build"

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

	ok, err := build.Push(
		pr.GetNumber(),
		pr.GetHead().GetSHA(),
		pr.GetHead().GetUser().GetLogin(),
	)
	if err != nil {
		return err
	}

	if ok {
		log.Println("Added to build queue")
	} else {
		log.Println("Not added to build queue")
	}

	return nil
}
