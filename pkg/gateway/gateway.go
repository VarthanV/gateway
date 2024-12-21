package gateway

import (
	"github.com/VarthanV/gateway/pkg/loadbalancer"
	"github.com/VarthanV/gateway/pkg/server"
)

type backend struct {
	servers []*server.Server
	lb      *loadbalancer.LoadBalancer
}

type gateway struct {
	Servers map[string]backend
}
