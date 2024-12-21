package middlewares

import (
	"net/http"
	"strings"

	"github.com/VarthanV/gateway/pkg/config"
)

type ServiceMiddleWare func(cfg *config.Config, serviceCfg *config.ServiceConfig, next http.HandlerFunc) http.HandlerFunc

func JWTMiddleware(cfg *config.Config, serviceCfg *config.ServiceConfig,
	next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.JWTConfig != nil && serviceCfg.JWTEnabled {
		}
	}
}

func CORSMiddleware(cfg *config.Config, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		origin := r.Header.Get("Origin")

		for _, o := range cfg.CORS.AllowedOrigins {
			if o == origin {
				w.Header().Set("Access-Control-Allow-Origin", o)
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.CORS.AllowedMethods, ","))
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
