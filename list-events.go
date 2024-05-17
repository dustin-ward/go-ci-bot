package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"test-org-gozbot/auth"
	"test-org-gozbot/config"
)

func main() {
	config.NewConfig(
		"https://github.ibm.com",
		"test-org-gozbot",
		"goz-workflow-demo",
		"./private-key.pem",
		2206,
		19342,
	)

	apiClient, err := auth.CreateClient()
	if err != nil {
		log.Fatal("Auth: ", err)
	}

	events, _, err := apiClient.Activity.ListRepositoryEvents(context.TODO(), config.Owner(), config.Repo(), nil)
	if err != nil {
		log.Fatal("ListRepositoryEvents: ", err)
	}
	if len(events) == 0 {
		fmt.Println("No Events")
		os.Exit(0)
	}

	for _, e := range events {
		fmt.Printf("Event: %s (%s)\n", e.GetType(), e.GetID())
		fmt.Println("  at:", e.GetCreatedAt())
		fmt.Println("  in:", e.GetRepo().GetName())
		fmt.Println("  by:", e.GetActor().GetLogin())
	}
}
