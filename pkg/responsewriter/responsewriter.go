package responsewriter

import (
	"bytes"
	"net/http"
)

type responseRecorder struct {
	http.ResponseWriter
	Body       *bytes.Buffer
	StatusCode int
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	rr.Body.Write(b)
	return rr.ResponseWriter.Write(b)
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.StatusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func NewResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		Body:           bytes.NewBuffer([]byte{}),
		StatusCode:     http.StatusOK,
	}
}
