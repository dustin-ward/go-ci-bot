package main

import (
	"flag"
	"fmt"
	"github.ibm.com/open-z/jeff-ci/config"
	"github.ibm.com/open-z/jeff-ci/gh"
	"log"
	"os"
)

func main() {
	n := flag.Int("n", 5, "Get last `n` events")
	flag.Parse()

	config.NewConfig(
	)

	events, err := gh.GetRepositoryEvents()
	if err != nil {
		log.Fatal(err)
	}
	if len(events) == 0 {
		fmt.Println("No Events")
		os.Exit(0)
	}

	for _, e := range events[:*n] {
		fmt.Printf("Event: %s (%s)\n", e.GetType(), e.GetID())
		fmt.Println("  at:", e.GetCreatedAt())
		fmt.Println("  in:", e.GetRepo().GetName())
		fmt.Println("  by:", e.GetActor().GetLogin())
	}
}
