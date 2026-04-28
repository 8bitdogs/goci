package payload

// README: https://developer.github.com/v3/repos/statuses/#create-a-status

type CommitStatusState string

const (
	// GitHub commit status states
	CommitStatusStateError   CommitStatusState = "error"
	CommitStatusStateFailure CommitStatusState = "failure"
	CommitStatusStatePending CommitStatusState = "pending"
	CommitStatusStateSuccess CommitStatusState = "success"
)

type StatusCreateRequest struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url,omitempty"`
	Description string `json:"description,omitempty"`
	Context     string `json:"context"`
}
