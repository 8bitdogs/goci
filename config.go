package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/8bitdogs/log"
	"github.com/antonmashko/envconf"
)

type config struct {
	CIHost   string `flag:"host" env:"CI_HOST"`
	CIFile   string `flag:"ci-config" default:"ci.json"`
	LogLevel string `env:"LOG_LEVEL" default:"info"`
	Github   struct {
		Secret string `env:"GITHUB_WEBHOOK_SECRET"`
		Token  string `env:"GITHUB_TOKEN"`
	}
	Server struct {
		Addr string `env:"SERVER_ADDR" default:":7878"`
	}
	CI []struct {
		WebhookPath string `json:"endpoint" required:"true"`
		Dir         string `json:"dir" required:"true"`
		Method      string `json:"method" default:"POST"`
	} `json:"ci"`
}

func parse() (*config, error) {
	var cfg config
	err := envconf.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(cfg.CIFile)
	if err != nil {
		return nil, err
	}
	lvl, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}
	log.DefaultLogger = log.New("goci", lvl)
	err = json.Unmarshal(b, &cfg.CI)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
