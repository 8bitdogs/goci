package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/8bitdogs/goci/core"
	"github.com/8bitdogs/log"
)

type Webhook struct {
	Timeout   time.Duration
	j         core.Job
	requestID uint64
}

func NewWebhook(j core.Job) *Webhook {
	return &Webhook{
		j:       j,
		Timeout: 8 * time.Second,
	}
}

func (wb *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var wp webhookPayload
	requestID := core.RequestID(r.Context())
	l := log.Copy(fmt.Sprintf("git-webhook-%d", requestID))
	err := json.NewDecoder(r.Body).Decode(&wp)
	if err != nil {
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
	go func() {
		l.Infoln("running job...")
		err := wb.j.Run(r.Context())
		lock <- err
	}()
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
	}
}
