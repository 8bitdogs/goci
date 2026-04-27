package option

var _ Option = (*tokenOption)(nil)

type tokenOption struct {
	token string
}

func (o *tokenOption) Apply(opts *Options) {
	opts.token = o.token
}

func WithToken(token string) Option {
	return &tokenOption{token: token}
}
