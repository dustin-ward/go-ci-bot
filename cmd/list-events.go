package main

import (
	"fmt"
	"log"
	"os"
	"test-org-gozbot/config"
	"test-org-gozbot/gh"
)

func main() {
	config.NewConfig(
		"https://github.ibm.com",
		"test-org-gozbot",
		"go",
		"./private-key.pem",
		2206,
		19342,
	)

	events, err := gh.GetRepositoryEvents()
	if err != nil {
		log.Fatal(err)
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
