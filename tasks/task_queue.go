package tasks

import (
	"sync"
)

var task_queue []Task
var mu sync.Mutex

func init() {
	task_queue = make([]Task, 0)
}

func Push(t Task) {
	mu.Lock()
	task_queue = append(task_queue, t)
	mu.Unlock()
}

func Pop() (Task, bool) {
	mu.Lock()
	defer mu.Unlock()

	if len(task_queue) > 0 {
		t := task_queue[0]
		task_queue = task_queue[1:]
		return t, true
	}

	return nil, false
}

func Poll() {
	task, ok := Pop()
	if ok {
		task.Do()
	}
}
