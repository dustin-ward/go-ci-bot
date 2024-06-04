package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"test-org-gozbot/config"
	"test-org-gozbot/events"
	"test-org-gozbot/tasks"
	"time"
)

const (
	EventPollInterval = time.Second * 10
	TaskPollInterval  = time.Second * 10
)

const (
	NumWorkers = 2
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

	// Wait for interrupt to end program
	stopMain := make(chan os.Signal, 1)
	signal.Notify(stopMain, os.Interrupt)

	// Goroutine communication
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Start worker goroutines
	for _ = range NumWorkers {
		worker := tasks.NewWorker()
		log.Printf("Starting Worker #%04d\n", worker.ID)

		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(TaskPollInterval)

			for {
				worker.Poll()

				// Wait for next poll or end.
				select {
				case <-ticker.C:
					continue
				case <-ctx.Done():
					// Stop worker
					return
				}
			}
		}()
	}

	// Poll for events
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(EventPollInterval)
		lastPollTime := time.Now() // Starting time
		log.Println("Polling for events...")

		for {
			// Wait for next poll or end.
			select {
			case <-ticker.C:
				// Do the poll
			case <-ctx.Done():
				// End Program
				return
			}

			// Events poll returns time right after poll was completed
			lastPollTime = events.Poll(lastPollTime, time.Now())
		}
	}()

	log.Println("GOZBOT has started")
	log.Println("Press Ctrl+C to exit")

	//TODO:Perform any takedown operations needed
	<-stopMain
	cancel()
	wg.Wait()
	log.Println("Shutting Down...")
}
