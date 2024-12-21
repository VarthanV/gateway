package server

import (
	"net/url"
	"sync/atomic"
)

type Server struct {
	url       *url.URL
	isHealthy atomic.Bool
}

func New(url string) *Server {
	return &Server{}
}

func (s *Server) SetHealth(val bool) {
	s.isHealthy.Store(val)
}

func (s *Server) GetHealth() bool {
	return s.isHealthy.Load()
}
