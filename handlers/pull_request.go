package handlers

import (
	"fmt"

	"github.com/google/go-github/v62/github"
)

func HandlePullRequestEvent(apiClient *github.Client, event *github.PullRequestEvent) error {
	fmt.Printf("Pull Request Event: #%d (%s)\n", event.GetPullRequest().GetNumber(), event.GetPullRequest().GetTitle())
	fmt.Println("  State:", event.GetPullRequest().GetState())
	fmt.Println("  Sender:", event.GetSender().GetLogin())
	fmt.Println("  Action:", event.GetAction())

	return nil
}
