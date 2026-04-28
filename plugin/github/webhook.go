package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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

	task := newCommitStatusPipelineTask(CommitStatusAPI{
		sha:         webhookRequest.Sha(),
		repository:  webhookRequest.FullName(),
		context:     wb.options.CommitStatusContext(),
		token:       wb.options.Token(),
		description: "goci pipeline",
		target_url:  "",
	}, wb.pipeline, l)

	_, err = wb.options.TaskQ().Enqueue(context.Background(), task)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		l.Error().Err(err).Msg("failed to enqueue pipeline task")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "pipeline triggered")
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
