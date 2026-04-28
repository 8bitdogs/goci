package github

import (
	"goci/plugin/github/option"

	"github.com/antonmashko/taskq"
)

type GithubConfig struct {
	Token string `yaml:"token" json:"token"`

	Webhook struct {
		CIHostURL           string `yaml:"ci_host" json:"ci_host"`
		Secret              string `yaml:"secret" json:"secret"`
		Method              string `yaml:"method" json:"method"`
		Path                string `yaml:"path" json:"path"`
		Branch              string `yaml:"branch" json:"branch"`
		EventType           string `yaml:"event" json:"event"`
		CommitStatusContext string `yaml:"commit_status_context" json:"commit_status_context"`
		Workflow            struct {
			Name    string `yaml:"name" json:"name"`
			Action  string `yaml:"action" json:"action"`
			JobName string `yaml:"job_name" json:"job_name"`
		} `yaml:"workflow" json:"workflow"`

		// ResponseTimeout time.Duration `yaml:"response_timeout" json:"response_timeout"`
	} `yaml:"webhook" json:"webhook"`
}

func (c *GithubConfig) Options(tq *taskq.TaskQ) []option.Option {
	return []option.Option{
		option.WithToken(c.Token),
		option.WithSecret(c.Webhook.Secret),
		option.WithEventType(c.Webhook.EventType),
		option.WithTargetBranch(c.Webhook.Branch),
		option.WithCommitStatusContext(c.Webhook.CommitStatusContext),
		option.WithWorkflowActionType(c.Webhook.Workflow.Action),
		option.WithWorkflowName(c.Webhook.Workflow.Name),
		option.WithWorkflowJobName(c.Webhook.Workflow.JobName),
		option.WithCIHostUrl(c.Webhook.CIHostURL),
		// option.WithTimeout(c.Webhook.ResponseTimeout),
		option.WithTaskQ(tq),
	}
}
