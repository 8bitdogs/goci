package option

var _ Option = (*workflowNameOption)(nil)

type workflowNameOption struct {
	workflowName string
}

func (o *workflowNameOption) Apply(opts *Options) {
	opts.workflowName = o.workflowName
}

func WithWorkflowName(workflowName string) Option {
	return &workflowNameOption{workflowName: workflowName}
}
