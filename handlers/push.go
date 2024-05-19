package handlers

import (
	"context"
	"fmt"
	"log"
	"test-org-gozbot/build"
	"test-org-gozbot/config"

	"github.com/google/go-github/v62/github"
)

func HandlePushEvent(apiClient *github.Client, event *github.PushEvent) error {
	prList, _, err := apiClient.PullRequests.List(context.TODO(), config.Owner(), config.Repo(), nil)
	if err != nil {
		return err
	}

	for _, pr := range prList {
		if "refs/heads/"+pr.GetHead().GetRef() == event.GetRef() {
			if event.GetSize() == 0 {
				// Sanity check
				return fmt.Errorf("no commits in push event %s/%s", event.GetRef(), event.GetHead())
			}

			headCommit := event.Commits[event.GetSize()-1]
			log.Printf("Push event in PR#%d %s - %s (%s)\n",
				pr.GetNumber(),
				pr.GetTitle(),
				headCommit.GetAuthor().GetLogin(),
				headCommit.GetSHA()[:6],
			)

			ok, err := build.Push(apiClient, pr.GetNumber(), headCommit.GetSHA(), headCommit.GetAuthor().GetLogin())
			if err != nil {
				return err
			}
			if ok {
				log.Println("Added to build queue")
			} else {
				log.Println("Not added to build queue")
			}
		}
	}

	return nil
}
