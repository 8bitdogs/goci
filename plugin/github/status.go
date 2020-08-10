package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

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

func createStatus(wp *webhookPayload, result StatusCreateRequest) error {
	const host = "https://api.github.com"
	//POST /repos/:owner/:repo/statuses/:sha
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/repos%s/statuses/%s", host, wp.Repository.FullName, wp.After)

	rs, err := http.Post(url, "application/json", strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	if rs.StatusCode < 200 || rs.StatusCode > 299 {
		return fmt.Errorf("invalid status code. url=%s status_code=%d", url, rs.StatusCode)
	}

	return nil
}
