package server

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	Addr        string
	Handler     http.Handler
	IdleTimeout time.Duration
}

func NewServer(addr string, handler http.Handler, idleTimeout time.Duration) *Server {
	return &Server{
		Addr:        addr,
		Handler:     handler,
		IdleTimeout: idleTimeout,
	}
}

func (s *Server) Listen() {
	srv := &http.Server{
		Addr:        s.Addr,
		Handler:     s.Handler,
		IdleTimeout: s.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("failed to start server")
		return
	}
}
