package main

import (
	"net/http"

	"github.com/8bitdogs/goci/core"
	"github.com/8bitdogs/goci/plugin/github"
	"github.com/8bitdogs/goci/plugin/pipeline"
	"github.com/8bitdogs/log"
	"github.com/8bitdogs/ruffe"
)

func main() {
	log.Info("starting application...")
	cfg, err := parse()
	if err != nil {
		log.Fatal(err)
	}
	server := core.NewServer(cfg.Server.Addr)
	for _, s := range cfg.CI {
		// serve webhook
		log.Infof("adding webhook: %s", s.WebhookPath)
		gitWh := github.NewWebhook(pipeline.New(s.Dir))
		server.Handle(s.WebhookPath, s.Method, gitWh)
	}
	if cfg.Github.Secret != "" {
		log.Infoln("adding github secret validation")
		gs := github.NewSecret(cfg.Github.Secret, cfg.Github.Key)
		server.Use(ruffe.HandlerFunc(func(ctx ruffe.Context) error {
			if !gs.Validate(ctx.Request()) {
				return ctx.Result(http.StatusForbidden, nil)
			}
			return nil
		}))
	}
	log.Infof("starting server on '%s'", cfg.Server.Addr)
	if err = server.ListenAndServe(); err != nil {
		log.Fatal("error on listen and serve. err=", err)
	}
}
