package checks

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/google/go-github/v50/github"
)

var (
	CONCLUSION_SUCCESS string = "success"
	CONCLUSION_FAILURE string = "failure"
	CONCLUSION_SKIPPED string = "skipped"
	CONCLUSION_NEUTRAL string = "neutral"
)

func Build() (output *github.CheckRunOutput, conclusion string) {
	title := "Results"
	text := "No issues here!"
	status := "Success"
	output = nil
	conclusion = CONCLUSION_SUCCESS

	// Start of check
	log.Println("Starting Build...")
	time.Sleep(4 * time.Second)

	cmd := exec.Command("./random")
	if err := cmd.Start(); err != nil {
		conclusion = CONCLUSION_FAILURE
		status = "Errored"
		text = fmt.Sprintf("cmd.Start: %v", err)
		goto FINISH
	}
	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			log.Println("random failed:", err)
			conclusion = CONCLUSION_FAILURE
			status = "Failed"
			text = "This test failed on purpose, re-run it :)"
		} else {
			conclusion = CONCLUSION_FAILURE
			status = "Errored"
            text = fmt.Sprintf("cmd.Wait: %v", err)
		}
	}

FINISH:
	summary := "check-summary: " + status
	output = &github.CheckRunOutput{Title: &title, Text: &text, Summary: &summary}
	return
}

func Test() bool {
	return true
}

func PAX() bool {
	return true
}

func Undef() (output *github.CheckRunOutput, conclusion string) {
	title := "Results"
	summary := "Unknown check name"
	text := "This shouldnt happen... @dustin-ward"
	return &github.CheckRunOutput{
		Title:   &title,
		Text:    &text,
		Summary: &summary,
	}, CONCLUSION_NEUTRAL
}
