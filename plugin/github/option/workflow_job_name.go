package option

var _ Option = (*workflowJobNameOption)(nil)

type workflowJobNameOption struct {
	workflowJobName string
}

func (o *workflowJobNameOption) Apply(opts *Options) {
	opts.workflowJobName = o.workflowJobName
}

func WithWorkflowJobName(workflowJobName string) Option {
	return &workflowJobNameOption{workflowJobName: workflowJobName}
}
