package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"net/http"
	"strings"

	"github.com/8bitdogs/log"
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
		log.Debugf("github-webhook: header %s not found", headerName)
		return false
	}
	hSignature := strings.TrimLeft(hv, prefix)
	signature, err := hex.DecodeString(hSignature)
	if err != nil {
		log.Debugf("github-webhook: failed to decode string %s err=%s", hSignature, err)
		return false
	}
	result := hmac.Equal(signature, s.mac.Sum(nil))
	s.mac.Reset()
	return result
}
