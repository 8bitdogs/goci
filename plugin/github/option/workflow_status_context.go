package option

type WorkflowStatusContextOption struct {
	workflowStatusContext string
}

func (o *WorkflowStatusContextOption) Apply(opts *Options) {
	opts.workflowStatusContext = o.workflowStatusContext
}

func WithWorkflowStatusContext(workflowStatusContext string) Option {
	return &WorkflowStatusContextOption{workflowStatusContext: workflowStatusContext}
}
