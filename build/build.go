package build

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"test-org-gozbot/checks"
	"test-org-gozbot/config"
	"time"

	"github.com/google/go-github/v62/github"
)

type Build struct {
	PR          int
	Branch      string
	SHA         string
	SubmittedBy string

	CheckRun *github.CheckRun
}

const (
	connString = "zosgo@zoscan59.pok.stglabs.ibm.com"
)

var (
	title             = "z/OS Build & Test"
	summaryInProgress = "In Progress..."
	summaryCompleted  = "Completed"
)

var (
	GithubUpdateInterval = time.Second * 30
)

func (b *Build) Start(apiClient *github.Client) {
	log.Printf("Doing build #%d/%s (%s) - %s\n", b.PR, b.Branch, b.SHA[:6], b.SubmittedBy)

	buildMachine := "zoscan59"

	body := "The build is now in progress. Machine: " + buildMachine
	var err error
	b.CheckRun, _, err = apiClient.Checks.UpdateCheckRun(context.TODO(), config.Owner(), config.Repo(),
		b.CheckRun.GetID(),
		github.UpdateCheckRunOptions{
			Name:   title,
			Status: &checks.STATUS_IN_PROGRESS,
			Output: &github.CheckRunOutput{Title: &title, Summary: &summaryInProgress, Text: &body},
		},
	)
	if err != nil {
		log.Fatal("Build: ", err)
	}

	output, ok := b.Do(apiClient)
	var conclusion string
	if ok {
		conclusion = checks.CONCLUSION_SUCCESS
	} else {
		conclusion = checks.CONCLUSION_FAILURE
	}
	log.Printf("Build Completed [%s] #%d/%s (%s) - %s\n", conclusion, b.PR, b.Branch, b.SHA[:6], b.SubmittedBy)

	body = "The build has completed. Output:\n" + output
	b.CheckRun, _, err = apiClient.Checks.UpdateCheckRun(context.TODO(), config.Owner(), config.Repo(),
		b.CheckRun.GetID(),
		github.UpdateCheckRunOptions{
			Name:        title,
			Status:      &checks.STATUS_COMPLETED,
			Conclusion:  &conclusion,
			CompletedAt: &github.Timestamp{time.Now()},
			Output:      &github.CheckRunOutput{Title: &title, Summary: &summaryCompleted, Text: &body},
		},
	)
	if err != nil {
		log.Fatal("Build: ", err)
	}
}

func (b *Build) Do(apiClient *github.Client) (output string, ok bool) {
	output = "```\n"
	var outputMu sync.Mutex
	ok = true
	cmd := exec.Command(
		"ssh",
		connString,
		fmt.Sprintf("~/gozbot-build-test.sh %s/%s %s",
			config.Owner(),
			config.Repo(),
			b.Branch,
		),
	)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	multi := io.MultiReader(stdout, stderr)

	err := cmd.Start()
	if err != nil {
		output += fmt.Sprintf("%v\n```", err)
		ok = false
		return
	}

	// Github Updater
	ctx, cancelUpdates := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(GithubUpdateInterval)
		for {
			select {
			case <-ticker.C:
				// Do update
			case <-ctx.Done():
				return
			}

			prevCheckRun := b.CheckRun
			outputMu.Lock()
			b.CheckRun, _, err = apiClient.Checks.UpdateCheckRun(context.TODO(), config.Owner(), config.Repo(),
				b.CheckRun.GetID(),
				github.UpdateCheckRunOptions{
					Name:   title,
					Status: &checks.STATUS_IN_PROGRESS,
					Output: &github.CheckRunOutput{Title: &title, Summary: &summaryInProgress, Text: &output},
				},
			)
			outputMu.Unlock()
			if err != nil {
				log.Println("Error updating github build status: ", err)
				b.CheckRun = prevCheckRun
			}
		}
	}()

	// Read from stdout+stderr
	scanner := bufio.NewScanner(multi)
	for scanner.Scan() {
		txt := scanner.Text()
		outputMu.Lock()
		output += "\n" + txt
		// Github CheckRun output limit is 65535 character.
		//TODO: Do something better than this
		if len(output) >= 65530 {
			over := len(output) - 65530
			output = output[over:]
		}
		outputMu.Unlock()
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	err = cmd.Wait()
	cancelUpdates()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			output += fmt.Sprintf("\nExit Code: %d", exitError.ExitCode())
		} else {
			output += fmt.Sprintf("\nUndefined Error: %v", err)
			log.Println("Build Error:", err)
		}
		ok = false
	}
	output += "\n```"

	return
}
