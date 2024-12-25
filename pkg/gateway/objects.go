package gateway

import "net/http"

type logInput struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Service        string
}
