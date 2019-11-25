package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/antonmashko/envconf"
)

type config struct {
	CIFile string `flag:"ci-config" json:"-" default:"ci.json"`
	Secret string `env:"SECRET" required:"true"`
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
	err = json.Unmarshal(b, &cfg.CI)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
