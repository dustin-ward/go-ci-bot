package build

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"test-org-gozbot/checks"
	"test-org-gozbot/config"
	"time"

	"github.com/google/go-github/v62/github"
)

type Build struct {
	PR          int
	SHA         string
	SubmittedBy string

	CheckRun *github.CheckRun
}

func (b *Build) Start(apiClient *github.Client) {
	log.Printf("Doing build #%d (%s) - %s\n", b.PR, b.SHA[:6], b.SubmittedBy)

	buildMachine := "zoscan56"

	title := "z/OS Build & Test"
	summary := "In Progress..."
	body := "The build is now in progress. Machine: " + buildMachine
	var err error
	b.CheckRun, _, err = apiClient.Checks.UpdateCheckRun(context.TODO(), config.Owner(), config.Repo(),
		b.CheckRun.GetID(),
		github.UpdateCheckRunOptions{
			Name:   title,
			Status: &checks.STATUS_IN_PROGRESS,
			Output: &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &body},
		},
	)
	if err != nil {
		log.Fatal("Build: ", err)
	}

	output, ok := b.Do()
	var conclusion string
	if ok {
		conclusion = checks.CONCLUSION_SUCCESS
	} else {
		conclusion = checks.CONCLUSION_FAILURE
	}
    log.Printf("Build Completed [%s] #%d (%s) - %s\n", conclusion, b.PR, b.SHA[:6], b.SubmittedBy)

	summary = "Completed"
	body = "The build has completed. Output:\n" + output
	b.CheckRun, _, err = apiClient.Checks.UpdateCheckRun(context.TODO(), config.Owner(), config.Repo(),
		b.CheckRun.GetID(),
		github.UpdateCheckRunOptions{
			Name:        title,
			Status:      &checks.STATUS_COMPLETED,
			Conclusion:  &conclusion,
			CompletedAt: &github.Timestamp{time.Now()},
			Output:      &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &body},
		},
	)
	if err != nil {
		log.Fatal("Build: ", err)
	}
}

func (b *Build) Do() (output string, ok bool) {
    time.Sleep(time.Second*10)
	n := rand.Intn(100)
	output = fmt.Sprintf("Random number was: %d\n", n)

	ok = n >= 25

	return
}
