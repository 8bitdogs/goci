package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"net/http"
)

type Secret struct {
	mac []byte
}

func NewSecret(secret, key string) *Secret {
	if key == "" {
		key = "sha1"
	}
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(secret))
	return &Secret{
		mac: mac.Sum(nil),
	}
}

func (s *Secret) Validate(r *http.Request) bool {
	const headerName = "X-Hub-Signature"
	hv := r.Header.Get(headerName)
	return hv != "" && hmac.Equal([]byte(hv), s.mac)
}

func (s *Secret) Handle(w http.ResponseWriter, r *http.Request) {
	if !s.Validate(r) {
		w.WriteHeader(http.StatusForbidden)
	}
}
