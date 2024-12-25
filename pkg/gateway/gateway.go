package gateway

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/VarthanV/gateway/pkg/config"
	gatewayerrors "github.com/VarthanV/gateway/pkg/gateway-errors"
	"github.com/VarthanV/gateway/pkg/loadbalancer"
	"github.com/VarthanV/gateway/pkg/log"
	"github.com/VarthanV/gateway/pkg/middlewares"
	"github.com/VarthanV/gateway/pkg/server"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type backend struct {
	servers      []*server.Server
	middlewares  []middlewares.MiddlewareFunc
	lb           *loadbalancer.LoadBalancer
	cfg          *config.ServiceConfig
	rateLimiters map[string]*rate.Limiter
}

type Gateway struct {
	servers       map[string]backend
	cfg           *config.Config
	logFile       *os.File
	logWriterChan chan logInput
}

func New(cfg *config.Config) *Gateway {

	var (
		maxLogWritersAllowed = 100
		writersSem           = make(chan struct{}, maxLogWritersAllowed)
	)

	g := Gateway{
		servers:       map[string]backend{},
		cfg:           cfg,
		logWriterChan: make(chan logInput), // Max writers allowed

	}

	go func() {
		logrus.Info("Logger worker started")
		for val := range g.logWriterChan {
			writersSem <- struct{}{}
			g.writeLog(val)
			<-writersSem
		}
	}()

	for _, c := range cfg.Services {
		b := backend{}
		for _, u := range c.Upstreams {
			b.servers = append(b.servers, server.New(u.URL))
		}

		b.lb = loadbalancer.New(loadbalancer.Algorithm(
			cfg.LoadBalancing.Algorithm))
		b.middlewares = append(b.middlewares, middlewares.DefaultMiddlewares...)
		b.cfg = &c
		b.rateLimiters = make(map[string]*rate.Limiter)
		g.servers[c.Path] = b
	}

	if cfg.Logging != nil {
		f, err := os.OpenFile(cfg.Logging.File,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logrus.Error("error in initializing log file ", err)
		}

		g.logFile = f
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
	g.logWriterChan <- logInput{
		Request:        r,
		ResponseWriter: w,
		Service:        urlSplit[0],
	}

}

func (g *Gateway) getServicePaths(path string) []string {
	re := regexp.MustCompile(`(\/[^/]+)(\/)`)
	result := re.ReplaceAllString(path, "$1|")

	// Split the result if needed
	parts := strings.Split(result, "|")
	return parts
}

func (g *Gateway) writeLog(i logInput) {

	l := log.Log{
		Service: i.Service,
		Path:    i.Request.URL.Path,
	}

	// Read the request body
	requestBody, err := io.ReadAll(i.Request.Body)
	if err != nil {
		logrus.Error("error reading request body: ", err)
		return
	}

	l.RequestBody = string(requestBody)

	// Read request headers
	requestHeaders, err := json.Marshal(i.Request.Header)
	if err != nil {
		logrus.Error("error marshalling request headers: ", err)
		return
	}
	l.RequestHeaders = string(requestHeaders)

	l.ResponseStatusCode = i.ResponseWriter.Header().Get("Status")
	responseHeaders, err := json.Marshal(i.ResponseWriter.Header())
	if err != nil {
		logrus.Error("error marshalling response headers: ", err)
		return
	}
	l.ResponseHeaders = string(responseHeaders)

	// Capture the response body if possible
	if rw, ok := i.ResponseWriter.(interface{ Body() []byte }); ok {
		l.ResponseBody = string(rw.Body())
	} else {
		l.ResponseBody = "response body not accessible"
	}

	// Marshal the log entry
	marshalledLog, err := json.Marshal(l)
	if err != nil {
		logrus.Error("error marshalling log: ", err)
		return
	}

	// Write log to file
	_, err = g.logFile.Write(append(marshalledLog, '\n'))
	if err != nil {
		logrus.Error("error writing log to file: ", err)
	}

	logrus.Info("Written log!")

}
