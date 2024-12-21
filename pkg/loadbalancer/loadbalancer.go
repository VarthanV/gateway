package loadbalancer

import (
	"errors"
	"math/rand"
	"sync/atomic"

	"github.com/VarthanV/gateway/pkg/server"
)

type Algorithm string

const (
	RoundRobin Algorithm = "round_robin"
	Random     Algorithm = "random"
)

type LoadBalancer struct {
	current   atomic.Int32
	algorithm Algorithm
}

func New(algo Algorithm) *LoadBalancer {
	return &LoadBalancer{algorithm: algo}
}

func (l *LoadBalancer) GetNextServer(servers []*server.Server) (*server.Server, error) {
	switch l.algorithm {
	case RoundRobin:
		return l.roundRobin(servers)
	case Random:
		return l.random(servers)

	default:
		return nil, errors.New("unimpleneted load balancing algorithm")
	}
}

func (l *LoadBalancer) roundRobin(servers []*server.Server) (*server.Server, error) {
	for i := 0; i < len(servers); i++ {
		idx := int(l.current.Load()) % len(servers)
		nextServer := servers[idx]
		l.current.Add(1)
		if nextServer.GetHealth() {
			return nextServer, nil
		}
	}

	return nil, errors.New("no server is healthy")

}

func (l *LoadBalancer) random(servers []*server.Server) (*server.Server, error) {
	idx := rand.Intn(len(servers) - 1)

	s := servers[idx]
	l.current.Add(1)

	if s.GetHealth() {
		return s, nil
	}

	// loop through and return the first healthy server
	for _, s := range servers {
		if s.GetHealth() {
			return s, nil
		}
	}

	return nil, errors.New("no server is healthy")

}
