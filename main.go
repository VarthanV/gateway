package main

import (
	"fmt"
	"net/http"

	"github.com/VarthanV/gateway/pkg/config"
	"github.com/sirupsen/logrus"
)

func main() {

	cfg := config.Config{}

	cfg.Load("config.toml")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "You accessed: %s", r.URL.Path)

	})

	logrus.Info("Starting gateway on port ", cfg.Server.Port)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Server.Host,
		cfg.Server.Port), nil)
	if err != nil {
		logrus.Fatal("unable to listen nd serve ", err)
	}
}
