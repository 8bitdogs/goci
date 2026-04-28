package pipeline

import (
	"bytes"
	"context"
	"goci/plugin/pipeline/cmd"

	"github.com/rs/zerolog/log"
)

type Runner struct {
	cfg *Config
}

func New(cfg *Config) *Runner {
	return &Runner{
		cfg: cfg,
	}
}

func (w *Runner) Run(ctx context.Context) error {
	l := log.With().
		Str("plugin", "pipeline").
		Logger()

	l.Info().Msg("Running pipeline plugin")
	stdOutBuff := bytes.NewBuffer(make([]byte, 1024*1024)) // 1MB buffer for stdout
	stdErrBuff := bytes.NewBuffer(make([]byte, 1024*1024)) // 1MB buffer for stderr
	for _, job := range w.cfg.Jobs {
		l.Info().Str("job", job.Name).Msg("Starting job")
		for _, step := range job.Steps {
			stdOutBuff.Reset()
			stdErrBuff.Reset()
			cmdStr := step.CmdString()
			l.Info().
				Str("job", job.Name).
				Str("step", step.Name).
				Str("cmd", cmdStr).
				Msg("Running step")
			c := cmd.NewCommand(step.Cmd, step.Args)
			if step.Dir != "" {
				c.SetDir(step.Dir)
			}
			c.SetStdout(stdOutBuff)
			c.SetStderr(stdErrBuff)
			if err := c.Run(ctx); err != nil {
				l.Error().
					Str("job", job.Name).
					Str("step", step.Name).
					Str("cmd", cmdStr).
					Str("stdout", stdOutBuff.String()).
					Str("stderr", stdErrBuff.String()).
					Err(err).
					Msg("Step failed")
				return err
			}
			l.Info().
				Str("job", job.Name).
				Str("step", step.Name).
				Str("cmd", cmdStr).
				Str("stdout", stdOutBuff.String()).
				Str("stderr", stdErrBuff.String()).
				Msg("Step complete")
		}
	}

	l.Info().Str("plugin", "pipeline").Msg("Pipeline plugin complete")
	return nil
}
