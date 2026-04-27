package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"goci/core"
	"goci/plugin/github/option"
	"goci/plugin/github/payload"

	"github.com/rs/zerolog/log"
)

type Webhook struct {
	pipeline  core.Pipeline
	signature *signature
	options   *option.Options
}

func NewWebhook(p core.Pipeline, opts ...option.Option) *Webhook {
	options := option.NewOptions(opts...)
	return &Webhook{
		pipeline:  p,
		options:   options,
		signature: newSignature(options.Secret()),
	}
}

func (wb *Webhook) createStatus(r payload.Request, result *payload.StatusCreateRequest) error {
	const host = "https://api.github.com"
	//POST /repos/:owner/:repo/statuses/:sha
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/repos/%s/statuses/%s", host, r.FullName(), r.Sha())

	l := log.With().
		Str("url", url).
		Str("status", result.State).
		Str("ci_url", result.TargetURL).
		Str("body", string(b)).
		Logger()

	l.Debug().Msg("sending status")

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(b)))
	if err != nil {
		l.Error().Err(err).Msg("failed to create request")
		return err
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprint("Bearer ", wb.options.Token()))
	rs, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Error().Err(err).Msg("failed to send status")
		return err
	}

	if rs.StatusCode < 200 || rs.StatusCode > 299 {
		err := fmt.Errorf("invalid status code. url=%s status_code=%d", url, rs.StatusCode)
		l.Error().Err(err).Msg("failed to send status")
		return err
	}

	l.Debug().Msg("Done")
	return nil
}

func (wb *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := core.RequestID(r.Context())
	l := log.With().
		Uint64("request_id", requestID).
		Logger()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		l.Error().Err(err).Msg("failed to read git payload")
		return
	}
	// validate signature
	if !wb.signature.validate(b, r) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(http.StatusText(http.StatusForbidden)))
		l.Warn().Msg("invalid signature")
		return
	}

	var webhookRequest payload.Request
	const eventHeader = "X-GitHub-Event"
	const contentTypeHeader = "Content-Type"
	eventType := r.Header.Get(eventHeader)
	if eventType == "" || !wb.options.IsEventType(eventType) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "skipped. expected %s got %s", wb.options.EventType(), eventType)
		l.Debug().Str("expected_event_type", wb.options.EventType()).Str("event_type", eventType).Msg("unsupported event type")
		return
	}

	contentType := r.Header.Get(contentTypeHeader)
	switch eventType {
	case "ping":
		p := &payload.Ping{}
		l.Debug().Msg("received ping request")
		if err := wb.unmarshalPayload(b, contentType, p); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
			l.Error().Err(err).Msg("failed to unmarshal ping payload")
			return
		}
		webhookRequest = p
	case "push":
		pp := &payload.Push{}
		if err := wb.unmarshalPayload(b, contentType, pp); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
			l.Error().Err(err).Msg("failed to unmarshal push payload")
			return
		}
		l.Debug().
			Str("ref", pp.Ref).
			Str("message", pp.HeadCommit.Message).
			Str("author", pp.HeadCommit.HeadCommitAuthor.Name).
			Str("branch", pp.Ref).
			Msg("received push request")
		webhookRequest = pp
	case "workflow_job":
		wj := &payload.WorkflowJob{}
		if err := wb.unmarshalPayload(b, contentType, wj); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
			l.Error().Err(err).Msg("failed to unmarshal workflow_job payload")
			return
		}

		if wb.options.WorkflowName() != "" && !wb.options.IsWorkflowName(wj.WorkflowJob.WorkflowName) {
			l.Debug().Str("workflow", wj.WorkflowJob.WorkflowName).Msg("skipped. workflow name does not match")
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprintf(w, "skipped. workflow name '%s' does not match target workflow name '%s'",
				wj.WorkflowJob.WorkflowName,
				wb.options.WorkflowName())
			return
		}

		if wb.options.WorkflowJobName() == "" || !wb.options.IsWorkflowJobName(wj.WorkflowJob.Name) {
			l.Debug().Str("job", wj.WorkflowJob.Name).Msg("skipped. workflow job name does not match")
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprintf(w, "skipped. workflow job name '%s' does not match target workflow job name '%s'",
				wj.WorkflowJob.Name,
				wb.options.WorkflowJobName())
			return
		}

		if wb.options.WorkflowAction() == "" || !wb.options.IsWorkflowAction(wj.Action) {
			l.Debug().Str("action", wj.Action).Msg("skipped. workflow action does not match")
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprintf(w, "skipped. workflow action '%s' does not match target workflow action '%s'",
				wj.Action,
				wb.options.WorkflowAction())
			return
		}

		l.Debug().
			Str("workflow", wj.WorkflowJob.WorkflowName).
			Str("job", wj.WorkflowJob.Name).
			Str("action", wj.Action).
			Msg("received workflow_job request")
		webhookRequest = wj
	default:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unsupported event type: %s", eventType)
		return
	}

	if !strings.EqualFold(wb.options.TargetBranch(), webhookRequest.TargetBranch()) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "skipped. branch '%s' does not match target branch '%s'",
			webhookRequest.TargetBranch(),
			wb.options.TargetBranch())
		return
	}

	lock := make(chan error)
	go func(data payload.Request) {
		l.Debug().Msg("running job...")
		err := wb.pipeline.Run(r.Context())

		if err != nil {
			l.Error().Err(err).Msg("run failed")
			select {
			case _, ok := <-lock:
				if !ok {
					break
				}
			default:
				lock <- err
			}
		}

		result := payload.StatusCreateRequest{
			State:       payload.Success.String(),
			Description: "",
			TargetURL:   "", // wb.options.CIHostUrl() + strconv.FormatUint(requestID, 10),
			Context:     wb.options.WorkflowStatusContext(),
		}

		if err != nil {
			result.State = payload.Error.String()
			result.Description = err.Error()
		}

		err = wb.createStatus(data, &result)
		if err != nil {
			l.Error().Err(err).Msg("failed to create github status")
		}
	}(webhookRequest)

	select {
	case err := <-lock:
		if err != nil {
			l.Error().Err(err).Msg("failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		l.Debug().Msg("job finished successful")
	case <-time.After(wb.options.Timeout()):
		w.Header().Add("X-Request-ID", fmt.Sprint(requestID))
		w.WriteHeader(http.StatusCreated)
		close(lock)
	}
}

func (wb *Webhook) unmarshalPayload(b []byte, contentType string, data any) error {
	switch contentType {
	case "application/json":
		return json.Unmarshal(b, data)
	case "application/x-www-form-urlencoded":
		values, err := url.ParseQuery(string(b))
		if err != nil {
			return err
		}
		payload := values.Get("payload")
		return json.Unmarshal([]byte(payload), data)
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
}
