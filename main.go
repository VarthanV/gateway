package main

import (
	"fmt"
	"net/http"

	"github.com/VarthanV/gateway/pkg/config"
	"github.com/VarthanV/gateway/pkg/gateway"
	"github.com/sirupsen/logrus"
)

func main() {

	cfg := config.Config{}

	cfg.Load("config.toml")
	logrus.Infof("Config is %+v", cfg)

	g := gateway.New(&cfg)

	mux := http.NewServeMux()
	g.RegisterRoutes(mux)

	logrus.Info("Starting gateway on port ", cfg.Server.Port)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Server.Host,
		cfg.Server.Port), mux)
	if err != nil {
		logrus.Fatal("unable to listen nd serve ", err)
	}
}
