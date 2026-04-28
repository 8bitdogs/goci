package github

import (
	"context"
	"fmt"
	"goci/core"
	"goci/plugin/github/payload"

	"github.com/antonmashko/taskq"
	"github.com/rs/zerolog"
)

var _ taskq.Task = (*commitStatusPipelineTask)(nil)
var _ taskq.TaskDone = (*commitStatusPipelineTask)(nil)
var _ taskq.TaskOnError = (*commitStatusPipelineTask)(nil)

type commitStatusPipelineTask struct {
	cs       CommitStatusAPI
	pipeline core.Pipeline
	log      zerolog.Logger
}

func newCommitStatusPipelineTask(cs CommitStatusAPI, pipeline core.Pipeline, log zerolog.Logger) *commitStatusPipelineTask {
	return &commitStatusPipelineTask{
		cs:       cs,
		pipeline: pipeline,
		log: log.With().
			Str("name", "commitStatusPipelineTask").
			Str("sha", cs.sha).
			Str("repository", cs.repository).
			Str("context", cs.context).
			Logger(),
	}
}

func (t *commitStatusPipelineTask) Do(ctx context.Context) error {
	t.log.Debug().Str("state", "pending").Msg("setting commit status")
	t.cs.state = payload.CommitStatusStatePending
	err := t.cs.Send(ctx)
	if err != nil {
		t.log.Error().Err(err).Msg("CommitStatusAPI.Send failed")
		return fmt.Errorf("CommitStatusAPI.Send: state=%s %w", t.cs.state, err)
	}
	t.log.Debug().Msg("running pipeline")
	err = t.pipeline.Run(ctx)
	if err != nil {
		t.log.Error().Err(err).Msg("pipeline.Run failed")
		return fmt.Errorf("pipeline.Run: %w", err)
	}
	return nil
}

func (t *commitStatusPipelineTask) Done(ctx context.Context) {
	t.log.Debug().Str("state", "success").Msg("task complete. setting commit status")
	t.cs.state = payload.CommitStatusStateSuccess
	err := t.cs.Send(ctx)
	if err != nil {
		// log error but do not retry
		t.log.Error().Err(err).Msgf("CommitStatusAPI.Send failed: state=%s", t.cs.state)
	}
}

func (t *commitStatusPipelineTask) OnError(ctx context.Context, err error) {
	t.log.Error().Str("state", "failure").Err(err).Msg("task failed. setting commit status")
	t.cs.state = payload.CommitStatusStateFailure
	sendErr := t.cs.Send(ctx)
	if sendErr != nil {
		// log error but do not retry
		t.log.Error().Err(sendErr).Msgf("CommitStatusAPI.Send failed: state=%s", t.cs.state)
	}
}
