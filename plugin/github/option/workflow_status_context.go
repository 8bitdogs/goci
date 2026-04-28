package option

var _ Option = (*CommitStatusContextOption)(nil)

type CommitStatusContextOption struct {
	commitStatusContext string
}

func (o *CommitStatusContextOption) Apply(opts *Options) {
	opts.commitStatusContext = o.commitStatusContext
}

func WithCommitStatusContext(commitStatusContext string) Option {
	return &CommitStatusContextOption{commitStatusContext: commitStatusContext}
}
