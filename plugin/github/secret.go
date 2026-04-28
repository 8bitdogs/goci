package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type signature struct {
	secret string
}

func newSignature(secret string) *signature {
	return &signature{secret: secret}
}

func (s *signature) validate(payload []byte, r *http.Request) bool {
	const headerName = "X-Hub-Signature"
	const prefix = "sha1="
	if s.secret == "" {
		return true
	}

	hv := r.Header.Get(headerName)
	if hv == "" {
		log.Debug().Str("header", headerName).Msg("github-webhook: header not found")
		return false
	}
	hSignature := strings.TrimPrefix(hv, prefix)
	signature, err := hex.DecodeString(hSignature)

	if err != nil {
		log.Debug().
			Str("header", headerName).
			Str("value", hSignature).
			Err(err).
			Msg("github-webhook: failed to decode string")
		return false
	}

	hm := hmac.New(sha1.New, []byte(s.secret))
	_, _ = hm.Write(payload)
	return hmac.Equal(signature, hm.Sum(nil))
}
