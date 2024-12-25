package main

import (
	"fmt"
	"net/http"

	"github.com/VarthanV/gateway/pkg/config"
	"github.com/VarthanV/gateway/pkg/gateway"
	"github.com/VarthanV/gateway/pkg/handlers"
	"github.com/VarthanV/gateway/pkg/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {

	cfg := config.Config{}

	cfg.Load("config.toml")
	logrus.Infof("Config is %+v", cfg)

	g := gateway.New(&cfg)

	corsWrappedHandler := middlewares.CORSMiddleware(&cfg,
		handlers.MainHandler(g))

	// Run the manager

	go func() {
		logrus.Info(`
			___   ___  ______  ____ __    __  ___  _  _
			// \\ // \\ | || | ||    ||    || // \\ \\//
			(( ___ ||=||   ||   ||==  \\ /\ // ||=||  )/ 
			\\_|| || ||   ||   ||___  \V/\V/  || || //  
														
	`)
		r := gin.Default()
		g.RegisterRoutes(r)
		r.Run(fmt.Sprintf(":%d", cfg.Server.ManagerPort))
	}()

	logrus.Info("Starting gateway on port ", cfg.Server.Port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Server.Host,
		cfg.Server.Port), corsWrappedHandler)
	if err != nil {
		logrus.Fatal("unable to listen nd serve ", err)
	}
}
