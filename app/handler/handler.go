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
	for _, tag := range req.Tags {
		if tag == model.TypeUser || tag == model.TypeOrganization {
			req.Timestamp = worker.OwnerWorker.Timestamp
			break
		}
		if tag == model.TypeRepository {
			req.Timestamp = worker.RepositoryWorker.Timestamp
			break
		}
	}

	var ranks []model.Rank
	model.NewRankModel().List(req, &ranks)

	response(w, http.StatusOK, ranks)
}

func response(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", http.MethodGet)
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(Payload{Data: payload}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func closeBody(r *http.Request) {
	if err := r.Body.Close(); err != nil {
		log.Fatal(err.Error())
	}
}
