package handler

import (
	"encoding/json"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/worker"
	"github.com/spf13/viper"
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

	timestamp := ""
	switch req.Type {
	case model.TypeUser:
		timestamp = worker.TimestampUserRanks
	case model.TypeOrganization:
		timestamp = worker.TimestampOrganizationRanks
	case model.TypeRepository:
		timestamp = worker.TimestampRepositoryRanks
	}

	ranks := worker.RankModel.List(req, time.Unix(0, viper.GetInt64(timestamp)))

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
