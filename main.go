package main

import (
	"goci/core"
	"goci/plugin/git"
	"goci/plugin/pipeline"

	"github.com/8bitdogs/log"
)

func main() {
	log.Info("starting application...")
	cfg, err := parse()
	if err != nil {
		log.Fatal(err)
	}
	server := core.NewServer(cfg.Server.Addr, cfg.Secret)
	for _, s := range cfg.CI {
		// serve webhook
		log.Infof("adding webhook: %s", s.WebhookPath)
		gitWh := git.NewWebhook(pipeline.New(s.Dir))
		server.Handle(s.WebhookPath, s.Method, gitWh)
	}
	if err = server.ListenAndServe(); err != nil {
		log.Fatal("error on listen and serve. err=", err)
	}
}
