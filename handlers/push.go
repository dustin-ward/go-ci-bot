package handlers

import (
	"fmt"

	"github.com/google/go-github/v62/github"
)

func HandlePushEvent(apiClient *github.Client, event *github.PushEvent) error {
	fmt.Printf("Push Event: %d\n", event.GetPushID())
	fmt.Println("\tHead:", event.GetHead())
	fmt.Println("\tRef:", event.GetRef())
	fmt.Println("\tSize:", event.GetSize())

	// if !buildqueue.RefInQueue(event.GetRef()) {
	//     buildqueue.Push(event)
	// }

	return nil

	// checkRun, _, err := apiClient.Checks.CreateCheckRun(context.TODO(), config.Owner(), config.Repo(),
	// 	github.CreateCheckRunOptions{
	// 		Name:      "z/OS Build",
	// 		HeadSHA:   event.GetHead(),
	// 		Status:    &checks.STATUS_QUEUED,
	// 		StartedAt: &github.Timestamp{time.Now()},
	// 	},
	// )
	// if err != nil {
	// 	log.Fatal("Create Check Run: ", err)
	// }

	// time.Sleep(15 * time.Second)

	// title := "TEST RUN"
	// summary := "This is the summary"
	// body := "Body text here... The run is now in progress"
	// checkRun, _, err = apiClient.Checks.UpdateCheckRun(context.TODO(), config.Owner(), config.Repo(),
	// 	checkRun.GetID(),
	// 	github.UpdateCheckRunOptions{
	// 		Name:   "z/OS Build",
	// 		Status: &checks.STATUS_IN_PROGRESS,
	// Output: &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &body},
	// 	},
	// )
	// if err != nil {
	// 	log.Fatal("Update Check Run 1: ", err)
	// }

	// time.Sleep(15 * time.Second)

	// body += "\n\nRun Completed"
	// checkRun, _, err = apiClient.Checks.UpdateCheckRun(context.TODO(), config.Owner(), config.Repo(),
	// 	checkRun.GetID(),
	// 	github.UpdateCheckRunOptions{
	// 		Name:        "z/OS Build",
	// 		Status:      &checks.STATUS_COMPLETED,
	// 		Conclusion:  &checks.CONCLUSION_FAILURE,
	// 		CompletedAt: &github.Timestamp{time.Now()},
	// Output: &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &body},
	// 	},
	// )
	// if err != nil {
	// 	log.Fatal("Update Check Run 2: ", err)
	// }

	// return nil
}
