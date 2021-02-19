package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/worker"
	"net/http"
)

var (
	rankModel         = model.NewRankModel()
	userModel         = model.NewUserModel()
	organizationModel = model.NewOrganizationModel()
	repositoryModel   = model.NewRepositoryModel()
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
	case app.TypeUser:
		req.Timestamps = append(req.Timestamps, worker.UserWorker.Timestamp)
	case app.TypeOrganization:
		req.Timestamps = append(req.Timestamps, worker.OrganizationWorker.Timestamp)
	case app.TypeRepository:
		req.Timestamps = append(req.Timestamps, worker.RepositoryWorker.Timestamp)
	default:
		req.Timestamps = append(req.Timestamps, worker.UserWorker.Timestamp)
		req.Timestamps = append(req.Timestamps, worker.OrganizationWorker.Timestamp)
		req.Timestamps = append(req.Timestamps, worker.RepositoryWorker.Timestamp)
	}

	ranks := rankModel.List(req)

	response(w, http.StatusOK, Payload{Data: ranks})
}

func ShowUser(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	user := userModel.FindByName(name)
	if user.ID() == "" {
		response(w, http.StatusNotFound, Payload{Data: nil})
		return
	}

	response(w, http.StatusOK, Payload{Data: user})
}

func ShowOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	organization := organizationModel.FindByName(name)
	if organization.ID() == "" {
		response(w, http.StatusNotFound, Payload{Data: nil})
		return
	}

	response(w, http.StatusOK, Payload{Data: organization})
}

func ShowRepository(w http.ResponseWriter, r *http.Request) {
	owner := mux.Vars(r)["owner"]
	name := mux.Vars(r)["name"]

	repository := repositoryModel.FindByName(fmt.Sprintf("%s/%s", owner, name))
	if repository.ID() == "" {
		response(w, http.StatusNotFound, Payload{Data: nil})
		return
	}

	response(w, http.StatusOK, Payload{Data: repository})
}

func response(w http.ResponseWriter, code int, payload Payload) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", http.MethodGet)
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
