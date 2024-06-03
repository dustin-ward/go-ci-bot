package tasks

import (
	"fmt"
	"log"
)

type Worker struct {
	ID         int
	CurTask    *Task
	CurMachine string
}

var nextId int
var workerPool map[int]*Worker

func init() {
	nextId = 0
	workerPool = make(map[int]*Worker)
}

func NewWorker() *Worker {
	id := nextId
	nextId++
	workerPool[id] = &Worker{id, nil, ""}
	return workerPool[id]
}

func (w *Worker) Poll() {
	workerStr := fmt.Sprintf("Worker #%04d", w.ID) // TODO: Give tasks id's (uuid/hash?)
	task, ok := popTask()
	if ok {
		host, ok := task.Provision()
		if !ok {
			log.Printf("%s: Unable to provision task %T\n", workerStr, task)
			task.Enqueue()
			return
		}

		log.Printf("%s: Starting %T on %s\n", workerStr, task, host)
		w.CurMachine = host
		w.CurTask = &task
		task.Do(host)
		w.CurTask = nil
		w.CurMachine = ""
		log.Printf("%s: Finished %T on %s\n", workerStr, task, host)
	}
}
