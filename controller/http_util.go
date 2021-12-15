package controller

import (
	"encoding/json"
	"net/http"
)

type Message struct {
	Message string `json:"message"`
}

func respondMessageJSON(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(Message{Message: msg})
}

func respondStructJSON(w http.ResponseWriter, body interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
