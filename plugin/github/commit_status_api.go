package github

import (
	"context"
	"encoding/json"
	"fmt"
	"goci/plugin/github/payload"
	"net/http"
	"strings"
)

// https://docs.github.com/en/rest/commits/statuses?apiVersion=2026-03-10#create-a-commit-status
type CommitStatusAPI struct {
	token       string                    // GitHub token with repo:status scope
	sha         string                    // commit SHA
	repository  string                    // format: owner/repo
	context     string                    // commit status context
	description string                    // short description of the commit status
	target_url  string                    // URL associated with the commit status
	state       payload.CommitStatusState // state of the commit status: "error", "failure", "pending", or "success"
}

func (cs *CommitStatusAPI) Send(ctx context.Context) error {
	const host = "https://api.github.com"
	payload := payload.StatusCreateRequest{
		State:       string(cs.state),
		TargetURL:   cs.target_url,
		Description: cs.description,
		Context:     cs.context,
	}

	//POST /repos/:owner/:repo/statuses/:sha
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/statuses/%s", host, cs.repository, cs.sha)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(b)))
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprint("Bearer ", cs.token))
	rs, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.DefaultClient.Do: url=%s body=%s %w", url, string(b), err)
	}

	if rs.StatusCode < 200 || rs.StatusCode > 299 {
		return fmt.Errorf("invalid status code. url=%s body=%s status_code=%d", url, string(b), rs.StatusCode)
	}

	return nil
}
