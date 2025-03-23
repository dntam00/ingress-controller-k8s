package server

import (
	"crypto/tls"
	"custom-ingress/gateway"
	"errors"
	"log"
	"net/http"
	"time"
)

type Server struct {
	Addr        string
	Handler     http.Handler
	IdleTimeout time.Duration

	k8sGateWay *gateway.Gateway
}

func NewServer(addr string, handler http.Handler, idleTimeout time.Duration, k8sGateway *gateway.Gateway) *Server {
	srv := &Server{
		Addr:        addr,
		Handler:     handler,
		IdleTimeout: idleTimeout,
		k8sGateWay:  k8sGateway,
	}
	return srv
}

func (s *Server) Listen() {
	srv := &http.Server{
		Addr:        s.Addr,
		Handler:     s.Handler,
		IdleTimeout: s.IdleTimeout,
	}

	srv.TLSConfig = &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert := s.k8sGateWay.GetRoute().GetCert(info.ServerName)
			if cert == nil {
				log.Printf("No certificate found for server %s", info.ServerName)
				return nil, errors.New("cert not found")
			}
			return cert, nil
		},
	}

	log.Println("Start listening on", s.Addr)

	if err := srv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Println("failed to start server")
		return
	}
}
