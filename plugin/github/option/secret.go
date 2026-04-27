package option

var _ Option = (*secretOption)(nil)

type secretOption struct {
	secret string
}

func (o *secretOption) Apply(opts *Options) {
	opts.secret = o.secret
}

func WithSecret(secret string) Option {
	return &secretOption{secret: secret}
}
