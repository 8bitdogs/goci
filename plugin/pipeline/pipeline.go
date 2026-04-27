package pipeline

import (
	"context"

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
	return nil
}
