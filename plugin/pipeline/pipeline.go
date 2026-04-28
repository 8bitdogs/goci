package pipeline

import (
	"context"
	"time"

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
	log.Info().Str("plugin", "pipeline").Msg("Running pipeline plugin")
	time.Sleep(30 * time.Second)
	log.Info().Str("plugin", "pipeline").Msg("Pipeline plugin complete")
	return nil
}
