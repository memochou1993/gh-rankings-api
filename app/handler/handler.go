package handler

import (
	"encoding/json"
	"net/http"
)

func response(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
