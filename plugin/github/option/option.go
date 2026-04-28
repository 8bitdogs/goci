package option

import (
	"time"

	"github.com/antonmashko/taskq"
)

type Option interface {
	Apply(opts *Options)
}

type Options struct {
	secret    string
	token     string
	ciHostURL string

	eventType    string
	targetBranch string

	commitStatusContext string

	workflowName    string
	workflowJobName string
	workflowAction  string

	timeout time.Duration
	taskq   *taskq.TaskQ
}

func NewOptions(opts ...Option) *Options {
	o := &Options{
		timeout: time.Second * 10,
	}
	for _, opt := range opts {
		opt.Apply(o)
	}
	return o
}

func (o *Options) Secret() string {
	return o.secret
}

func (o *Options) Token() string {
	return o.token
}

func (o *Options) CIHostURL() string {
	return o.ciHostURL
}

func (o *Options) EventType() string {
	return o.eventType
}

func (o *Options) IsEventType(eventType string) bool {
	return o.eventType == eventType
}

func (o *Options) TargetBranch() string {
	return o.targetBranch
}

func (o *Options) IsTargetBranch(targetBranch string) bool {
	return o.targetBranch == targetBranch
}

func (o *Options) Timeout() time.Duration {
	return o.timeout
}

func (o *Options) WorkflowName() string {
	return o.workflowName
}

func (o *Options) IsWorkflowName(workflowName string) bool {
	return o.workflowName == workflowName
}

func (o *Options) WorkflowJobName() string {
	return o.workflowJobName
}

func (o *Options) IsWorkflowJobName(workflowJobName string) bool {
	return o.workflowJobName == workflowJobName
}

func (o *Options) WorkflowAction() string {
	return o.workflowAction
}

func (o *Options) IsWorkflowAction(workflowAction string) bool {
	return o.workflowAction == workflowAction
}

func (o *Options) CommitStatusContext() string {
	return o.commitStatusContext
}

func (o *Options) TaskQ() *taskq.TaskQ {
	return o.taskq
}
