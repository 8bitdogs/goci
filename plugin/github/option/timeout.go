package option

import "time"

var _ Option = (*timeoutOption)(nil)

type timeoutOption struct {
	timeout time.Duration
}

func (o *timeoutOption) Apply(opts *Options) {
	opts.timeout = o.timeout
}

func WithTimeout(timeout time.Duration) Option {
	return &timeoutOption{timeout: timeout}
}
