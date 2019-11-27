package main

import (
	"github.com/8bitdogs/goci/core"
	"github.com/8bitdogs/goci/plugin/git"
	"github.com/8bitdogs/goci/plugin/pipeline"
	"github.com/8bitdogs/log"
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
		gitWh := git.NewWebhook(pipeline.New(s.Dir))
		server.Handle(s.WebhookPath, s.Method, gitWh)
	}
	if cfg.Secret != "" {
		server.ValidateSecret(cfg.Secret)
		log.Infoln("added secret validation")
	}
	if err = server.ListenAndServe(); err != nil {
		log.Fatal("error on listen and serve. err=", err)
	}
}
