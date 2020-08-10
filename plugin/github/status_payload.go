package github

// README: https://developer.github.com/v3/repos/statuses/#create-a-status

type Status byte

func (s Status) String() string {
	switch s {
	case Success:
		return "success"
	case Pending:
		return "pending"
	case Failure:
		return "failure"
	case Error:
		return "error"
	default:
		return ""
	}
}

const (
	Success Status = 0 + iota
	Pending
	Error
	Failure
)

type StatusCreateRequest struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url"`
	Description string `json:"description"`
	Context     string `json:"context"`
}
