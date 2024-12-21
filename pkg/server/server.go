package server

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

type Server struct {
	serverURL *url.URL
	isHealthy atomic.Bool
}

func New(serverURL string) *Server {
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		logrus.Error("error in parsing url ", err)
	}
	return &Server{
		serverURL: parsedURL,
	}
}

func (s *Server) SetHealth(val bool) {
	s.isHealthy.Store(val)
}

func (s *Server) GetHealth() bool {
	return s.isHealthy.Load()
}

func (s *Server) GetURL() *url.URL {
	return s.serverURL
}

func (s *Server) ReverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.GetURL())
}
