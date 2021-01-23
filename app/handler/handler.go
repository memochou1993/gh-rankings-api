package handler

import (
	"encoding/json"
	"github.com/memochou1993/github-rankings/app/handler/request"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/app/worker"
	"log"
	"net/http"
)

type Payload struct {
	Data interface{} `json:"data"`
}

func Index(w http.ResponseWriter, r *http.Request) {
	defer closeBody(r)

	req := request.Parse(r)
	switch {
	case req.HasTag(model.TypeUser):
		req.Timestamp = worker.OwnerWorker.Timestamp
	case req.HasTag(model.TypeOrganization):
		req.Timestamp = worker.OwnerWorker.Timestamp
	case req.HasTag(model.TypeRepository):
		req.Timestamp = worker.RepositoryWorker.Timestamp
	}

	ranks := model.NewRankModel().List(req)

	response(w, http.StatusOK, Payload{ranks})
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
