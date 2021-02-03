package handler

import (
	"encoding/json"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/worker"
	"log"
	"net/http"
	"time"
)

type Payload struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func Index(w http.ResponseWriter, r *http.Request) {
	defer closeBody(r)

	req, err := request.Validate(r)
	if err != nil {
		response(w, http.StatusUnprocessableEntity, Payload{Error: err.Error()})
		return
	}
	if req.Name == "" && req.Type == "" {
		response(w, http.StatusBadRequest, Payload{})
		return
	}
	timestamps := []time.Time{
		worker.OwnerWorker.Timestamp,
		worker.RepositoryWorker.Timestamp,
	}
	ranks := worker.RankModel.List(req, timestamps)

	response(w, http.StatusOK, Payload{Data: ranks})
}

func response(w http.ResponseWriter, code int, payload Payload) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", http.MethodGet)
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func closeBody(r *http.Request) {
	if err := r.Body.Close(); err != nil {
		log.Fatal(err.Error())
	}
}
