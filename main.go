package main

import (
	"os"

	"goci/core"
	"goci/plugin/github"
	"goci/plugin/pipeline"

	"github.com/8bitdogs/ruffe"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, serviceConfigs, err := parse()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse configuration")
	}

	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(cfg.Log.Level)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	server := core.NewServer(cfg.Server.Addr)

	// adding webhook handlers
	for _, s := range serviceConfigs {
		log.Info().Str("method", s.Github.Webhook.Method).
			Str("path", s.Github.Webhook.Path).
			Str("target_branch", s.Github.Webhook.Branch).
			Str("event_type", s.Github.Webhook.EventType).
			Str("workflow_name", s.Github.Webhook.Workflow.Name).
			Str("workflow_action", s.Github.Webhook.Workflow.Action).
			Str("workflow_job_name", s.Github.Webhook.Workflow.JobName).
			Str("ci_host", s.Github.Webhook.CIHostURL).
			// Dur("response_timeout", s.Github.Webhook.ResponseTimeout).
			Msg("adding webhook")

		gitWh := github.NewWebhook(pipeline.New(&s.Pipeline), s.Github.Options()...)
		server.Handle(s.Github.Webhook.Path, s.Github.Webhook.Method, ruffe.HTTPHandlerFunc(gitWh.ServeHTTP))
	}

	log.Info().Str("address", cfg.Server.Addr).Msg("starting server")
	if err = server.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("error on listen and serve")
	}
}
