package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type signature struct {
	mac hash.Hash
}

func newSignature(secret string) *signature {
	if secret == "" {
		return &signature{}
	}
	return &signature{
		mac: hmac.New(sha1.New, []byte(secret)),
	}
}

func (s *signature) validate(payload []byte, r *http.Request) bool {
	const headerName = "X-Hub-Signature"
	const prefix = "sha1="
	if s.mac == nil {
		return true
	}
	s.mac.Write(payload)
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
	result := hmac.Equal(signature, s.mac.Sum(nil))
	s.mac.Reset()
	return result
}
