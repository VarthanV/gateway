package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"

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

	s := &Server{
		serverURL: parsedURL,
	}

	s.healthCheck()
	go func() {
		for range time.Tick(5 * time.Second) {
			s.healthCheck()
		}
	}()
	return s
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

func (s *Server) healthCheck() {
	url := s.serverURL.String()

	logrus.Info("Health checking ", url)
	res, err := http.Head(url)
	if err != nil {
		logrus.Error("error in health check ", err)
		s.isHealthy.Store(false)
	}

	if res != nil && res.StatusCode >= http.StatusOK &&
		res.StatusCode < http.StatusBadRequest {
		s.isHealthy.Store(true)
	} else {
		s.isHealthy.Store(false)
	}

}
