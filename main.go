package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"test-org-gozbot/build"
	"test-org-gozbot/config"
	"test-org-gozbot/events"
	"time"
)

const (
	EventPollInterval = time.Second * 10
	BuildPollInterval = time.Second * 1
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

	// Cancel any checks that were previously queued
	log.Println("Cleaning Stale Events...")
    err := events.Cleanup()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("GOZBOT has started")

	// Wait for interrupt to end program
	stopMain := make(chan os.Signal, 1)
	signal.Notify(stopMain, os.Interrupt)
	log.Println("Press Ctrl+C to exit")

	// Goroutine communication
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Start build poller
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(BuildPollInterval)

		for {
			build.Poll()

			// Wait for next poll or end.
			// Do this after so that the first poll happens right on init instead of 60 seconds after
			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				// End Program
				return
			}
		}
	}()

	// Poll for events
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(EventPollInterval)
		lastPollTime := time.Now() // Starting time

		for {
			// Events poll returns time right after poll was completed
			lastPollTime = events.Poll(lastPollTime)

			// Wait for next poll or end.
			// Do this after so that the first poll happens right on init instead of 60 seconds after
			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				// End Program
				return
			}
		}
	}()

	//TODO:Perform any takedown operations needed
	<-stopMain
	cancel()
	wg.Wait()
	log.Println("Shutting Down...")
}
