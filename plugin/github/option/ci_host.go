package option

import "strings"

var _ Option = (*ciHostUrlOption)(nil)

type ciHostUrlOption struct {
	ciHostUrl string
}

func (o *ciHostUrlOption) Apply(opts *Options) {
	opts.ciHostUrl = o.ciHostUrl
}

// WithCIHostUrl sets the CI host URL to which the plugin will send build status updates.
// formats supported: http://ci.example.com, https://ci.example.com, ci.example.com
func WithCIHostUrl(ciHostUrl string) Option {
	ciHostUrl = strings.TrimRight(strings.TrimSpace(ciHostUrl), "/") + "/"
	return &ciHostUrlOption{ciHostUrl: ciHostUrl}
}
