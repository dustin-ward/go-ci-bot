package tasks

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"test-org-gozbot/config"
	"test-org-gozbot/gh"
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
	summaryInQueue    = "In Queue..."
	summaryInProgress = "In Progress..."
	summaryCompleted  = "Completed"
)

var (
	GithubUpdateInterval = time.Second * 30
)

func (b Build) Do() error {
	log.Printf("Doing build #%d/%s (%s) - %s\n", b.PR, b.Branch, b.SHA[:6], b.SubmittedBy)

	buildMachine := "zoscan59"

	body := "The build is now in progress. Machine: " + buildMachine
	var err error
	b.CheckRun, err = gh.UpdateCheckRun(b.CheckRun, summaryInProgress, body)
	if err != nil {
		return fmt.Errorf("Error Starting Build: %v", err)
	}

	output, ok := b.build()
	var conclusion string
	if ok {
		conclusion = gh.CHECK_CONCLUSION_SUCCESS
	} else {
		conclusion = gh.CHECK_CONCLUSION_FAILURE
	}
	log.Printf("Build Completed [%s] #%d/%s (%s) - %s\n", conclusion, b.PR, b.Branch, b.SHA[:6], b.SubmittedBy)

	body = "The build has completed. Output:\n" + output
	b.CheckRun, err = gh.CompleteCheckRun(b.CheckRun, conclusion, summaryCompleted, body)
	if err != nil {
		return fmt.Errorf("Error Concluding Build: %v", err)
	}

	return nil
}

func (b *Build) build() (output string, ok bool) {
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
			b.CheckRun, err = gh.UpdateCheckRun(b.CheckRun, summaryInProgress, output)
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

func PushBuild(PR int, SHA, SubmittedBy string) (ok bool, err error) {
	ok = false

	// Make sure PR is mergeable first
	pr, err := gh.GetPullRequest(PR)
	if err != nil {
		return
	}
	if !pr.GetMergeable() {
		msg := "⚠️ PR is not mergeable. Please resolve conflicts"
		_, err = gh.CreateComment(pr.GetNumber(), msg)
		return
	}

	// See if any builds for this SHA are already queued
	for _, t := range task_queue {
		if build, ok := t.(Build); ok && build.SHA == SHA {
			return false, nil
		}
	}

	msg := "This commit has been added to the build queue. More information will appear here once the build has started"
	checkRun, err := gh.CreateCheckRun(SHA, title, summaryInQueue, msg)
	if err != nil {
		return
	}

	Push(Build{
		PR:          PR,
		Branch:      pr.GetHead().GetRef(),
		SHA:         SHA,
		SubmittedBy: SubmittedBy,
		CheckRun:    checkRun,
	})
	if err != nil {
		return
	}

	ok = true
	return
}
