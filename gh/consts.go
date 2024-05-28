package gh

// Need to take the address of these string literals to provide
// them as args. So we assign them to variables first.
var (
	CHECK_STATUS_QUEUED      string = "queued"
	CHECK_STATUS_IN_PROGRESS string = "in_progress"
	CHECK_STATUS_COMPLETED   string = "completed"
)

var (
	CHECK_CONCLUSION_SUCCESS         string = "success"
	CHECK_CONCLUSION_FAILURE         string = "failure"
	CHECK_CONCLUSION_NEUTRAL         string = "neutral"
	CHECK_CONCLUSION_CANCELLED       string = "cancelled"
	CHECK_CONCLUSION_SKIPPED         string = "skipped"
	CHECK_CONCLUSION_TIMED_OUT       string = "timed_out"
	CHECK_CONCLUSION_ACTION_REQUIRED string = "action_required"
)
