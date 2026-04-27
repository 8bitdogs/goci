package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"goci/plugin/github"
	"goci/plugin/pipeline"

	"github.com/antonmashko/envconf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type config struct {
	CIHost     string `flag:"host" env:"CI_HOST"`
	ConfigFile string `flag:"config" default:"config.json"`

	Server struct {
		Addr string `env:"SERVER_ADDR" default:":7878"`
	}

	Log struct {
		Level zerolog.Level `env:"LOG_LEVEL" default:"info"`
	}

	Github struct {
		Token                 string        `env:"GITHUB_TOKEN"`
		Method                string        `env:"GITHUB_METHOD" default:"POST"`
		ResponseTimeout       time.Duration `env:"GITHUB_RESPONSE_TIMEOUT" default:"10s"`
		Secret                string        `env:"GITHUB_WEBHOOK_SECRET"`
		TargetBranch          string        `env:"GITHUB_TARGET_BRANCH" default:"main"`
		EventType             string        `env:"GITHUB_EVENT_TYPE" default:"push"`
		WorkflowName          string        `env:"GITHUB_WORKFLOW_NAME"`
		WorkflowJobName       string        `env:"GITHUB_WORKFLOW_JOB_NAME"`
		WorkflowAction        string        `env:"GITHUB_WORKFLOW_ACTION" default:"completed"`
		WorkflowStatusContext string        `env:"GITHUB_WORKFLOW_STATUS_CONTEXT" default:"deploy"`
	}
}

type serviceConfig struct {
	Name     string              `yaml:"name" json:"name"`
	Github   github.GithubConfig `yaml:"github" json:"github"`
	Pipeline pipeline.Config     `yaml:"pipeline" json:"pipeline"`
}

func parse() (*config, []serviceConfig, error) {
	var cfg config
	err := envconf.Parse(&cfg)
	if err != nil {
		return nil, nil, err
	}

	ext := filepath.Ext(cfg.ConfigFile)

	b, err := os.ReadFile(cfg.ConfigFile)
	if err != nil {
		return nil, nil, err
	}

	var serviceCfg []serviceConfig
	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(b, &serviceCfg)
		if err != nil {
			return nil, nil, err
		}
	case ".json":
		err = json.Unmarshal(b, &serviceCfg)
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, fmt.Errorf("unsupported config file format: %s. use one of .yaml, .yml, .json", ext)
	}

	for _, s := range serviceCfg {
		if s.Github.Webhook.Method == "" {
			if cfg.Github.Method != "" {
				s.Github.Webhook.Method = cfg.Github.Method
			} else {
				s.Github.Webhook.Method = http.MethodPost
			}
		}

		if s.Github.Webhook.Secret == "" {
			s.Github.Webhook.Secret = cfg.Github.Secret
		}

		if s.Github.Token == "" {
			s.Github.Token = cfg.Github.Token
		}

		if s.Github.Webhook.Branch == "" {
			s.Github.Webhook.Branch = cfg.Github.TargetBranch
		}

		// if s.Github.Webhook.ResponseTimeout == 0 {
		// 	s.Github.Webhook.ResponseTimeout = cfg.Github.ResponseTimeout
		// }

		log.Info().Str("service", s.Name).Str("event", s.Github.Webhook.EventType).Msg("parsing service configuration")
		if s.Github.Webhook.EventType == "" {
			if cfg.Github.EventType != "" {
				s.Github.Webhook.EventType = cfg.Github.EventType
			} else {
				s.Github.Webhook.EventType = "push"
			}
		}

		if cfg.Github.WorkflowName != "" && s.Github.Webhook.Workflow.Name == "" {
			s.Github.Webhook.Workflow.Name = cfg.Github.WorkflowName
		}

		if s.Github.Webhook.Workflow.JobName == "" && strings.HasPrefix(s.Github.Webhook.EventType, "workflow") {
			if cfg.Github.WorkflowJobName != "" {
				s.Github.Webhook.Workflow.JobName = cfg.Github.WorkflowJobName
			} else {
				panic("`job_name` required for workflow events")
			}
		}

		if s.Github.Webhook.Workflow.Action == "" && strings.HasPrefix(s.Github.Webhook.EventType, "workflow") {
			if cfg.Github.WorkflowAction != "" {
				s.Github.Webhook.Workflow.Action = cfg.Github.WorkflowAction
			} else {
				panic("`action` required for workflow events")
			}
		}

		if s.Github.Webhook.Workflow.StatusContext == "" && strings.HasPrefix(s.Github.Webhook.EventType, "workflow") {
			if cfg.Github.WorkflowStatusContext != "" {
				s.Github.Webhook.Workflow.StatusContext = cfg.Github.WorkflowStatusContext
			} else {
				panic("`status_context` required for workflow events")
			}
		}
	}

	return &cfg, serviceCfg, nil
}
