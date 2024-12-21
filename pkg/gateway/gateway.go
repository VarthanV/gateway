package gateway

import (
	"github.com/VarthanV/gateway/pkg/config"
	"github.com/VarthanV/gateway/pkg/loadbalancer"
	"github.com/VarthanV/gateway/pkg/server"
)

type backend struct {
	methods   []string
	servers   []*server.Server
	lb        *loadbalancer.LoadBalancer
	stripPath bool
}

type gateway struct {
	servers map[string]backend
}

func New(cfg *config.Config) *gateway {
	g := gateway{
		servers: map[string]backend{},
	}

	for _, r := range cfg.Routes {
		b := backend{}
		for _, u := range r.Upstreams {
			b.servers = append(b.servers, server.New(u.URL))
		}

		b.lb = loadbalancer.New(loadbalancer.Algorithm(
			cfg.LoadBalancing.Algorithm))
		b.methods = r.Methods
		b.stripPath = r.StripPath
		g.servers[r.Path] = b
	}

	return &g
}
