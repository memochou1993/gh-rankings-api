package handler

import (
	"encoding/json"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/worker"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

type Payload struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func Index(w http.ResponseWriter, r *http.Request) {
	defer app.CloseBody(r.Body)

	req, err := request.Validate(r)
	if err != nil {
		response(w, http.StatusUnprocessableEntity, Payload{Error: err.Error()})
		return
	}
	if req.Name == "" && req.Type == "" {
		response(w, http.StatusBadRequest, Payload{})
		return
	}

	var ranks []model.Rank
	switch req.Type {
	case model.TypeUser:
		ranks = worker.NewUserWorker().RankModel.List(req, time.Unix(0, viper.GetInt64(worker.TimestampUserRanks)))
	case model.TypeOrganization:
		ranks = worker.NewOrganizationWorker().RankModel.List(req, time.Unix(0, viper.GetInt64(worker.TimestampOrganizationRanks)))
	case model.TypeRepository:
		ranks = worker.NewRepositoryWorker().RankModel.List(req, time.Unix(0, viper.GetInt64(worker.TimestampRepositoryRanks)))
	}

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
