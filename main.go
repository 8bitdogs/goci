package main

import (
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
	// adding webhook handlers
	for _, s := range cfg.CI {
		log.Infof("adding webhook: %s", s.WebhookPath)
		gitWh := github.NewWebhook(pipeline.New(s.Dir), cfg.Github.Secret)
		server.Handle(s.WebhookPath, s.Method, ruffe.HTTPHandlerFunc(gitWh.ServeHTTP))
	}
	log.Infof("starting server on '%s'", cfg.Server.Addr)
	if err = server.ListenAndServe(); err != nil {
		log.Fatal("error on listen and serve. err=", err)
	}
}
