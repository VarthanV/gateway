package handlers

import (
	"net/http"

	"github.com/VarthanV/gateway/pkg/gateway"
)

func MainHandler(g *gateway.Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		g.HandleRequest(w, r)
	}

}
