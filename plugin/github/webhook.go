package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/8bitdogs/goci/core"
	"github.com/8bitdogs/log"
)

type Webhook struct {
	Timeout   time.Duration
	j         core.Job
	signature *signature
	token     string
}

func NewWebhook(j core.Job, secret, token string) *Webhook {
	return &Webhook{
		j:         j,
		Timeout:   8 * time.Second,
		signature: newSignature(secret),
		token:     token,
	}
}

func (wb *Webhook) createStatus(wp *webhookPayload, result StatusCreateRequest) error {
	const host = "https://api.github.com"
	//POST /repos/:owner/:repo/statuses/:sha
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/repos/%s/statuses/%s", host, wp.Repository.FullName, wp.After)

	l := log.Copy(fmt.Sprintf("url=%s status=%s ci_url=%s", url, result.State, result.TargetURL))
	l.Infoln("sendign status")

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprint("token ", wb.token))
	rs, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	if rs.StatusCode < 200 || rs.StatusCode > 299 {
		return fmt.Errorf("invalid status code. url=%s status_code=%d", url, rs.StatusCode)
	}

	l.Infoln("Done")
	return nil
}

func (wb *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := core.RequestID(r.Context())
	l := log.Copy(fmt.Sprintf("git-webhook-%d", requestID))
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		l.Errorf("failed to read git payload. err=%s", err)
		return
	}
	// validate signature
	if !wb.signature.validate(b, r) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	var wp webhookPayload
	if err = json.Unmarshal(b, &wp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		l.Errorf("failed to unmarshal git payload. err=%s", err)
		return
	}
	l.Infof("received git webhook. message=%s author=%s branch=%s", wp.HeadCommit.Message, wp.HeadCommit.Author, wp.Ref)
	const master = "refs/heads/master"
	if !strings.EqualFold(master, wp.Ref) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Payload not for master, aborting"))
		return
	}
	lock := make(chan error)
	go func(payload *webhookPayload) {
		l.Infoln("running job...")
		err := wb.j.Run(r.Context())

		if err != nil {
			l.Errorf("run failed. error=%s", err)
			select {
			case _, ok := <-lock:
				if !ok {
					break
				}
			default:
				lock <- err
			}
		}

		result := StatusCreateRequest{
			State:       Success.String(),
			Description: "",
			TargetURL:   fmt.Sprintf("http://ci.jared.in.ua/%d", requestID),
			Context:     "8bitdogs/goci",
		}

		if err != nil {
			result.State = Error.String()
			result.Description = err.Error()
		}

		err = wb.createStatus(payload, result)
		if err != nil {
			l.Errorf("failed to create github status. err=%s", err)
		}
	}(&wp)
	select {
	case err := <-lock:
		if err != nil {
			l.Infof("failed. err=%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		l.Infof("job finished successful")
	case <-time.After(wb.Timeout):
		w.Header().Add("X-Request-ID", fmt.Sprint(requestID))
		w.WriteHeader(http.StatusCreated)
		close(lock)
	}
}
