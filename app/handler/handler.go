package handler

import (
	"encoding/json"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/worker"
	"net/http"
)

var (
	RankModel = model.NewRankModel()
)

type Payload struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func Index(w http.ResponseWriter, r *http.Request) {
	defer app.CloseBody(r.Body)

	req, err := request.New(r)
	if err != nil {
		response(w, http.StatusUnprocessableEntity, Payload{Error: err.Error()})
		return
	}

	switch req.Type {
	case model.TypeUser:
		req.Timestamps = append(req.Timestamps, worker.UserWorker.Timestamp)
	case model.TypeOrganization:
		req.Timestamps = append(req.Timestamps, worker.OrganizationWorker.Timestamp)
	case model.TypeRepository:
		req.Timestamps = append(req.Timestamps, worker.RepositoryWorker.Timestamp)
	default:
		req.Timestamps = append(req.Timestamps, worker.UserWorker.Timestamp)
		req.Timestamps = append(req.Timestamps, worker.UserWorker.Timestamp)
		req.Timestamps = append(req.Timestamps, worker.UserWorker.Timestamp)
	}

	ranks := RankModel.List(req)

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
