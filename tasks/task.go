package tasks

type Task interface {
	// Task entry point. Do the actual work here
	Do() error
}
