package core

import (
	"net/http"
	"sync/atomic"

	"github.com/8bitdogs/ruffe"
	"github.com/rs/zerolog/log"
)

type Server struct {
	requestID uint64
	addr      string
	secret    string
	router    *ruffe.Router
}

func NewServer(addr string) *Server {
	router := ruffe.New()
	server := &Server{
		addr:   addr,
		router: router,
	}

	// add request id
	router.AppendInterceptor(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		r = r.WithContext(storeRequestID(r.Context(), atomic.AddUint64(&server.requestID, 1)))
		next(w, r)
	})

	// logging
	router.AppendInterceptor(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		log.Info().
			Uint64("request_id", RequestID(r.Context())).
			Str("method", r.Method).
			Str("request_uri", r.RequestURI).
			Str("proto", r.Proto).
			Interface("header", r.Header).
			Msg("incoming request")
		next(w, r)
	})

	return server
}

func (s *Server) Handle(pattern, method string, h ruffe.Handler) {
	s.router.Handle(pattern, method, h)
}

func (s *Server) Use(h ruffe.Handler) {
	s.router.Use(h)
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.router)
}
