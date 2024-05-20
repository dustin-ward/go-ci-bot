package build

import (
	"context"
	"log"
	"sync"
	"test-org-gozbot/checks"
	"test-org-gozbot/config"
	"time"

	"github.com/google/go-github/v62/github"
)

var build_queue []Build
var mu sync.Mutex

func init() {
	build_queue = make([]Build, 0)
}

func Push(apiClient *github.Client, PR int, SHA, SubmittedBy string) (ok bool, err error) {
	mu.Lock()
	defer mu.Unlock()
	ok = false

	// Make sure PR is mergeable first
	pr, _, err := apiClient.PullRequests.Get(context.TODO(), config.Owner(), config.Repo(), PR)
	if err != nil {
		return
	}
	if !pr.GetMergeable() {
		msg := "⚠️ PR is not mergeable. Please resolve conflicts"
		_, _, err = apiClient.Issues.CreateComment(context.TODO(), config.Owner(), config.Repo(), PR,
			&github.IssueComment{Body: &msg},
		)
		return
	}

	// See if any builds for this SHA are already queued
	for _, build := range build_queue {
		if build.SHA == SHA {
			return
		}
	}

	title := "z/OS Build & Test"
	summary := "In Queue"
	msg := "This commit has been added to the build queue. More information will appear here once the build has started"
	checkRun, _, err := apiClient.Checks.CreateCheckRun(context.TODO(), config.Owner(), config.Repo(),
		github.CreateCheckRunOptions{
			Name:      title,
			HeadSHA:   SHA,
			Status:    &checks.STATUS_QUEUED,
			StartedAt: &github.Timestamp{time.Now()},
			Output:    &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &msg},
		},
	)
	if err != nil {
		return
	}

	build_queue = append(build_queue, Build{
		PR:          PR,
		Branch:      pr.GetHead().GetRef(),
		SHA:         SHA,
		SubmittedBy: SubmittedBy,
		CheckRun:    checkRun,
	})

	ok = true
	return
}

func Pop() (build *Build, ok bool) {
	mu.Lock()
	defer mu.Unlock()
	build = nil
	ok = false

	// If any builds later in the queue exist for this PR, we should
	// cancel the current build and pop the next build instead.
	for len(build_queue) > 0 {
		build = &build_queue[0]
		build_queue = build_queue[1:]
		ok = true

		for _, nextBuild := range build_queue {
			if nextBuild.PR == build.PR {
				log.Printf("Skipped #%d (%s) because of newer build in queue (%s) \n", build.PR, build.SHA[:6], nextBuild.SHA[:6])
				build = nil
				ok = false
				break
			}
		}
		if ok {
			break
		}
	}

	return
}
