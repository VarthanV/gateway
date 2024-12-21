package gatewayerrors

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	ErrorKey = "error"
)

type Error struct {
	HttpStatusCode int         `json:"-"`
	Code           string      `json:"code"`
	Message        string      `json:"message"`
	AdditionalInfo interface{} `json:"additional_info"`
}

func Write(e *Error, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(e.HttpStatusCode)

	marshalledRes, err := json.Marshal(e)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.Header.Set(ErrorKey, e.Message)

	logrus.Debug("writing error ", marshalledRes)

	w.Write(marshalledRes)
}
