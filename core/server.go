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
	router    *ruffe.Router
}

func NewServer(addr string) *Server {
	router := ruffe.New()
	server := &Server{addr: addr}

	// add request id
	router.AppendInterceptor(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		r = r.WithContext(storeRequestID(r.Context(), atomic.AddUint64(&server.requestID, 1)))
		next(w, r)
	})

	// logging
	router.AppendInterceptor(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		log.Infof("request-%d %s %s %s %v", RequestID(r.Context()), r.Method, r.RequestURI, r.Proto, r.Header)
	})

	return server
}

func (s *Server) Handle(patter, method string, h ruffe.Handler) {
	s.router.Handle(patter, method, h)
}

func (s *Server) Use(h ruffe.Handler) {
	s.router.Use(h)
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.router)
}
