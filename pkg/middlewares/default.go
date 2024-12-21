package middlewares

import (
	"net/http"

	"github.com/VarthanV/gateway/pkg/config"
)

func RequestTaggingMiddleware(cfg *config.Config,
	serviceCfg *config.ServiceConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
