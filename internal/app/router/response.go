package router

import (
	"net/http"

	"demo_bot/pkg/marshaler"
)

type response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func WriteSuccess(w http.ResponseWriter, r any) {
	writeResponse(w, http.StatusOK, response{Data: r})
}

func WriteError(w http.ResponseWriter, status int, err error) {
	writeResponse(w, status, response{Error: err.Error()})
}

func writeResponse(w http.ResponseWriter, status int, r any) {
	b, err := marshaler.MarshalJSONWithoutEscape(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(b); err != nil {
		return
	}
}
