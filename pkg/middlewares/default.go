package middlewares

import (
	"net/http"
	"slices"

	"github.com/VarthanV/gateway/pkg/config"
	"github.com/VarthanV/gateway/pkg/constants"
	gatewayerrors "github.com/VarthanV/gateway/pkg/gateway-errors"
	"github.com/google/uuid"
)

type MiddlewareFunc func(serverConfig *config.Config, serviceConfig *config.ServiceConfig,
	w http.ResponseWriter, r *http.Request)

func RequestTaggingMiddleware(
	serverConfig *config.Config, serviceConfig *config.ServiceConfig,
	w http.ResponseWriter, r *http.Request) {
	id := uuid.NewString()
	r.Header.Set(constants.RequestIDHeader, id)
	w.Header().Set(constants.RequestIDHeader, id)

}

func CheckIfMethodAllowed(serverConfig *config.Config, serviceConfig *config.ServiceConfig,
	w http.ResponseWriter, r *http.Request) {
	if !slices.Contains(serviceConfig.Methods, r.Method) {
		gatewayerrors.Write(&gatewayerrors.Error{
			HttpStatusCode: http.StatusForbidden,
			Message:        "Method not allowed"}, w, r)
		return
	}
}

var DefaultMiddlewares = []MiddlewareFunc{
	RequestTaggingMiddleware,
	CheckIfMethodAllowed}
