package gateway

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/VarthanV/gateway/pkg/config"
	gatewayerrors "github.com/VarthanV/gateway/pkg/gateway-errors"
	"github.com/VarthanV/gateway/pkg/loadbalancer"
	"github.com/VarthanV/gateway/pkg/middlewares"
	"github.com/VarthanV/gateway/pkg/server"
	"github.com/sirupsen/logrus"
)

type backend struct {
	servers     []*server.Server
	middlewares []middlewares.MiddlewareFunc
	lb          *loadbalancer.LoadBalancer
	cfg         *config.ServiceConfig
}

type Gateway struct {
	servers map[string]backend
	cfg     *config.Config
}

func New(cfg *config.Config) *Gateway {
	g := Gateway{
		servers: map[string]backend{},
		cfg:     cfg,
	}

	for _, c := range cfg.Services {
		b := backend{}
		for _, u := range c.Upstreams {
			b.servers = append(b.servers, server.New(u.URL))
		}

		b.lb = loadbalancer.New(loadbalancer.Algorithm(
			cfg.LoadBalancing.Algorithm))
		b.middlewares = append(b.middlewares, middlewares.DefaultMiddlewares...)
		b.cfg = &c

		g.servers[c.Path] = b
	}

	return &g
}

func (g *Gateway) applyMiddlewares(b *backend, w http.ResponseWriter, r *http.Request) {
	for _, m := range b.middlewares {
		m(g.cfg, b.cfg, w, r)
	}
}

func (g *Gateway) HandleRequest(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("You accessed: %s", r.URL.Path)
	urlSplit := g.getServicePaths(r.URL.Path)
	if len(urlSplit) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid path"))
		return
	}

	b, ok := g.servers[urlSplit[0]]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	server, err := b.lb.GetNextServer(b.servers)
	if err != nil {
		logrus.Error("error in getting server ", err)
		gatewayerrors.Write(
			&gatewayerrors.Error{HttpStatusCode: 400, Message: "No healthy server found"}, w, r)
		return
	}

	if b.cfg.StripPath {
		originalPath := r.URL.Path
		trimmedPath := strings.TrimPrefix(originalPath, "/"+urlSplit[0])
		if !strings.HasPrefix(trimmedPath, "/") {
			trimmedPath = "/" + trimmedPath
		}
		r.URL.Path = trimmedPath
	}

	g.applyMiddlewares(&b, w, r)

	// FIXME: dirty workaround need to find a better way
	if r.Header.Get(gatewayerrors.ErrorKey) != "" {
		logrus.Error("error in request ", r.Header.Get(gatewayerrors.ErrorKey))
		r.Header.Del(gatewayerrors.ErrorKey)
		return
	}

	server.ReverseProxy().ServeHTTP(w, r)
}

func (g *Gateway) getServicePaths(path string) []string {
	re := regexp.MustCompile(`(\/[^/]+)(\/)`)
	result := re.ReplaceAllString(path, "$1|")

	// Split the result if needed
	parts := strings.Split(result, "|")
	return parts
}
