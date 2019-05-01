package amznode

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func respond(w http.ResponseWriter, r *http.Request, v interface{}, code int) {
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(v)
	if err != nil {
		respondErr(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_, err = b.WriteTo(w)
	if err != nil {
		log.Printf("respond: %s", err)
	}
}

func respondErr(w http.ResponseWriter, r *http.Request, err error, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	err = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
	if err != nil {
		log.Printf("respond: %s", err)
	}
}

// ErrorResponse is used to serialize errors to json
type ErrorResponse struct {
	Error string `json:"error"`
}
