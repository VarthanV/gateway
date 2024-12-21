package handlers

import (
	"fmt"
	"net/http"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "You accessed: %s", r.URL.Path)
}
