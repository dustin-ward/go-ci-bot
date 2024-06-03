package tasks

type Task interface {
	// Add this task to the global task_queue. Any github CheckRuns should be
	// created here too if needed.
	// If this task should no longer exist in the queue, it is the job of this
	// function to determine that and act as a no-op.
	Enqueue()

	// Execute logic to determine which machine this task will run on
	//   - The string value will be the hostname ('local' if not an ssh
	//     connection)
	//   - The boolean value represents a successful provision job. Returning
	//     false means that `Enqueue()` will be called again.
	Provision() (string, bool)

	// Task entry point. Do the actual work here
	//   - 'host' is the hostname to perform the task at. It could be empty or
	//     'local' to indicate that the field should be ignored.
	Do(host string) error
}
