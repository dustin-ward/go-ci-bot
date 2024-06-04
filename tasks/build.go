package tasks

import (
	"bufio"
	"context"
	"fmt"
	"github.ibm.com/open-z/jeff-ci/config"
	"github.ibm.com/open-z/jeff-ci/gh"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/google/go-github/v62/github"
)

type Build struct {
	PR          int
	Branch      string
	SHA         string
	SubmittedBy string

	BuildMachine string
	CheckRun     *github.CheckRun
}

var (
	buildTitle             = "z/OS Build & Test"
	buildSummaryInQueue    = "In Queue..."
	buildSummaryInProgress = "In Progress..."
	buildSummaryCompleted  = "Completed"
)

func (b Build) Enqueue() {
	// Make sure PR is mergeable first. Otherwise send a comment
	pr, err := gh.GetPullRequest(b.PR)
	if err != nil {
		log.Printf("Build not queued: %v\n", err)
		return
	}
	if !pr.GetMergeable() {
		msg := "⚠️ PR is not mergeable. Please resolve conflicts"
		_, err = gh.CreateComment(pr.GetNumber(), msg)
		log.Printf("Build not queued: PR not mergeable %v\n", err)
		return
	}

	// See if any workers are already doing this task
	for _, worker := range workerPool {
		if worker.CurTask != nil {
			if e, ok := (*worker.CurTask).(Build); ok && e.SHA == b.SHA {
				log.Printf("Build not queued: A worker is already proccessing this task\n")
				return
			}
		}
	}

	// See if any builds for this SHA are already queued
	for _, t := range task_queue {
		if build, ok := t.(Build); ok && build.SHA == b.SHA {
			log.Printf("Build not queued: This task is already in the queue\n")
			return
		}
	}

	// If this is a new task in the queue, create the initial github check status object
	if b.CheckRun == nil {
		msg := "This commit has been added to the build queue. More information will appear here once the build has started"
		b.CheckRun, err = gh.CreateCheckRun(b.SHA, buildTitle, buildSummaryInQueue, msg)
		if err != nil {
			log.Printf("Build not queued: error creating check run: %v\n", err)
			return
		}
	}

	pushTask(b)

	log.Printf("Build added to queue")
	return
}

func (b Build) Provision() (string, bool) {
	return getZosMachine()
}

func (b Build) Do(host string) error {
	buildStr := fmt.Sprintf("%s/%s [#%d] (%s) - %s", config.Repo(), b.Branch, b.PR, b.SHA[:6], b.SubmittedBy)
	log.Printf("Starting build %s\n", buildStr)

	// Update the github status to show "In-Progress"
	body := "The build is now in progress. Machine: " + host
	var err error
	b.CheckRun, err = gh.UpdateCheckRun(b.CheckRun, buildSummaryInProgress, body)
	if err != nil {
		return fmt.Errorf("Error Starting Build: %v", err)
	}

	// Do the SSH portion of the task
	output, ok := b.build(host)
	var conclusion string
	if ok {
		conclusion = gh.CHECK_CONCLUSION_SUCCESS
	} else {
		conclusion = gh.CHECK_CONCLUSION_FAILURE
	}
	log.Printf("Build Completed <%s> %s\n", conclusion, buildStr)

	// Update the github status to show "Completed" (PASS/FAIL)
	body = fmt.Sprintf("The build has completed. Output:\n```\n%s\n```", output)
	b.CheckRun, err = gh.CompleteCheckRun(b.CheckRun, conclusion, buildSummaryCompleted, body)
	if err != nil {
		return fmt.Errorf("Error Concluding Build: %v", err)
	}

	return nil
}

var (
	GithubUpdateInterval = time.Second * 30
)

func (b *Build) build(host string) (output string, ok bool) {
	var outputMu sync.Mutex
	ok = true

	// This is the script that will be executed on z/OS.
	// We need to make sure that the machines are actually setup to invoke this script.
	// Need:
	//   ~/gozbot-build-test.sh
	//   ~/gozbot/
	cmd := exec.Command(
		"ssh",
		fmt.Sprintf("zosgo@%s", host),
		fmt.Sprintf("~/gozbot-build-test.sh %s/%s %s",
			config.Owner(),
			config.Repo(),
			b.Branch,
		),
	)

	// Combine stdout and stderr so we can read from the combined pipe dynamically
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	multi := io.MultiReader(stdout, stderr)

	err := cmd.Start()
	if err != nil {
		output += err.Error()
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
			b.CheckRun, err = gh.UpdateCheckRun(b.CheckRun, buildSummaryInProgress, fmt.Sprintf("```\n%s\n```", output))
			outputMu.Unlock()
			if err != nil {
				log.Println("Error updating github build status: ", err)
				b.CheckRun = prevCheckRun
			}
		}
	}()

	// Read from the combined pipe. It will return EOF when the cmd finishes, so we can block here
	scanner := bufio.NewScanner(multi)
	for scanner.Scan() {
		txt := scanner.Text()
		outputMu.Lock()
		output += "\n" + txt
		// Github CheckRun output limit is 65535 character.
		//TODO: Do something better than this
		if len(output) >= 65000 {
			over := len(output) - 65000
			output = output[over:]
		}
		outputMu.Unlock()
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	// The command has finished. Now cleanup
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

	return
}
