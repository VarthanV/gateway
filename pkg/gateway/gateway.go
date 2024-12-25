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
	"github.com/VarthanV/gateway/pkg/responsewriter"
	"github.com/VarthanV/gateway/pkg/server"
	"github.com/gin-gonic/gin"
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
		gatewayerrors.Write(&gatewayerrors.Error{
			Message:        "Invalid path",
			HttpStatusCode: 400,
		}, w, r)
		return
	}

	defer func(w http.ResponseWriter, r *http.Request) {
		g.logWriterChan <- logInput{
			Request:        r,
			ResponseWriter: w,
			Service:        urlSplit[0],
		}
	}(w, r)

	b, ok := g.servers[urlSplit[0]]
	if !ok {
		gatewayerrors.Write(&gatewayerrors.Error{
			Message:        "Service not available",
			HttpStatusCode: 400,
		}, w, r)
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
	recorder := responsewriter.NewResponseRecorder(i.ResponseWriter)

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
	l.ResponseBody = recorder.Body.String()

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

func (g *Gateway) RegisterRoutes(r *gin.Engine) {
	r.POST("/servers", g.addServer)
	r.PUT("/servers", g.updateServer)
	r.DELETE("/servers", g.deleteServer)
	r.GET("/servers", g.listServers)
	r.GET("/logs", g.listLogs)
}

func (g *Gateway) addServer(c *gin.Context) {
	var newServer server.Server
	if err := c.ShouldBindJSON(&newServer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Add server logic here
	path := c.DefaultQuery("path", "")
	if backend, ok := g.servers[path]; ok {
		backend.servers = append(backend.servers, &newServer)
		g.servers[path] = backend
		c.JSON(http.StatusCreated, gin.H{"message": "Server added"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service path not found"})
	}
}

func (g *Gateway) deleteServer(c *gin.Context) {
	serverURL := c.DefaultQuery("url", "")
	path := c.DefaultQuery("path", "")
	if path == "" || serverURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing path or server URL"})
		return
	}

	if backend, ok := g.servers[path]; ok {
		for i, srv := range backend.servers {
			if srv.GetURL().String() == serverURL {
				backend.servers = append(backend.servers[:i], backend.servers[i+1:]...)
				g.servers[path] = backend
				c.JSON(http.StatusOK, gin.H{"message": "Server deleted"})
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service path not found"})
	}
}

func (g *Gateway) updateServer(c *gin.Context) {
	var updatedServer server.Server
	if err := c.ShouldBindJSON(&updatedServer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	path := c.DefaultQuery("path", "")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing path"})
		return
	}

	if backend, ok := g.servers[path]; ok {
		for i, srv := range backend.servers {
			if srv.GetURL() == updatedServer.GetURL() {
				backend.servers[i] = &updatedServer
				g.servers[path] = backend
				c.JSON(http.StatusOK, gin.H{"message": "Server updated"})
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service path not found"})
	}
}

func (g *Gateway) listServers(c *gin.Context) {
	path := c.DefaultQuery("path", "")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing path"})
		return
	}

	if backend, ok := g.servers[path]; ok {
		servers := make([]string, len(backend.servers))
		for i, srv := range backend.servers {
			servers[i] = srv.GetURL().String()
		}
		c.JSON(http.StatusOK, servers)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service path not found"})
	}
}

func (g *Gateway) listLogs(c *gin.Context) {
	if g.logFile == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Log file not initialized"})
		return
	}

	logs, err := os.ReadFile(g.logFile.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read log file"})
		return
	}

	c.Data(http.StatusOK, "application/json", logs)
}
