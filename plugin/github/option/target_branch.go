package option

var _ Option = (*targetBranchOption)(nil)

type targetBranchOption struct {
	targetBranch string
}

func (o *targetBranchOption) Apply(opts *Options) {
	opts.targetBranch = o.targetBranch
}

func WithTargetBranch(targetBranch string) Option {
	return &targetBranchOption{targetBranch: targetBranch}
}
