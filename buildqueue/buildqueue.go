package buildqueue

import (
	"fmt"
	"log"

	"github.com/google/go-github/v62/github"
)

type buildq_t struct {
	commitSHA string
    ref string
}

var build_queue []buildq_t

func Init(init_cap int) {
	build_queue = make([]buildq_t, 0, init_cap)
}

func Push(event *github.PushEvent) {
	build_queue = append(build_queue, buildq_t{event.GetHead(), event.GetRef()})
    log.Printf("Push Added to Build Queue: %s#%s\n", event.GetRef(), event.GetHead())
    fmt.Println(build_queue)
}

func Pop() string {
	if len(build_queue) == 0 {
		return ""
	}
    commitSHA := build_queue[0].commitSHA
	build_queue = build_queue[1:]
	log.Println("Removed from Build Queue:", commitSHA)
	return commitSHA
}

func RefInQueue(ref string) bool {
	for _, bq := range build_queue {
		if bq.ref == ref {
			return true
		}
	}
	return false
}
