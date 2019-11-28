package github

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
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
	Timeout time.Duration
	j       core.Job
	secret  string
}

func NewWebhook(j core.Job, secret string) *Webhook {
	return &Webhook{
		j:       j,
		Timeout: 8 * time.Second,
		secret:  secret,
	}
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
	if wb.secret != "" {
		if !wb.validateSignature(b, r) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}
	//-----
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

func (wb *Webhook) validateSignature(payload []byte, r *http.Request) bool {
	mac := hmac.New(sha1.New, []byte(wb.secret))
	mac.Write(payload)
	const headerName = "X-Hub-Signature"
	const prefix = "sha1="
	hv := r.Header.Get(headerName)
	signature := mac.Sum(nil)
	hs := bytes.TrimLeft([]byte(hv), prefix)
	log.Debugf("comparing signature %s[%s] sha1=%x", hv, signature)
	return hv != "" && hmac.Equal(hs, signature)
}
