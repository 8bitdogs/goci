package option

import "time"

type Option interface {
	Apply(opts *Options)
}

type Options struct {
	secret    string
	token     string
	ciHostUrl string

	eventType    string
	targetBranch string

	workflowName          string
	workflowJobName       string
	workflowAction        string
	workflowStatusContext string

	timeout time.Duration
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

func (o *Options) CIHostUrl() string {
	return o.ciHostUrl
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

func (o *Options) WorkflowStatusContext() string {
	return o.workflowStatusContext
}
