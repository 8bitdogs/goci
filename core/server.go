package core

import (
	"net/http"
	"sync/atomic"

	"github.com/8bitdogs/log"
	"github.com/8bitdogs/ruffe"
)

type Server struct {
	requestID uint64
	addr      string
	secret    string
	server    *ruffe.Server
}

func NewServer(addr string, secret string) *Server {
	return &Server{
		addr:   addr,
		server: ruffe.New(),
	}
}

func (s *Server) Handle(patter, method string, h http.Handler) {
	s.server.Handle(patter, method, ruffe.HTTPHandlerFunc(h.ServeHTTP))
}

func (s *Server) ListenAndServe() error {
	// logging
	s.server.Use(ruffe.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("request %s %s %s %v", r.Method, r.RequestURI, r.Proto, r.Header)
	}))
	// validate secret
	s.server.Use(ruffe.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))
	// add request id
	s.server.Use(ruffe.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*r = *r.WithContext(storeRequestID(r.Context(), atomic.AddUint64(&s.requestID, 1)))
	}))
	return http.ListenAndServe(s.addr, s.server)
}
