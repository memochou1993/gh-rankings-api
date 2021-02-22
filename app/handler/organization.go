package handler

import (
	"github.com/gorilla/mux"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/model"
	"net/http"
)

var (
	organizationModel = model.NewOrganizationModel()
)

func ListOrganizations(w http.ResponseWriter, r *http.Request) {
	defer app.CloseBody(r.Body)

	req, err := request.NewOrganizationRequest(r)
	if err != nil {
		response(w, http.StatusUnprocessableEntity, Payload{Error: err.Error()})
		return
	}

	organizations := organizationModel.List(req)

	response(w, http.StatusOK, Payload{Data: organizations})
}

func ShowOrganization(w http.ResponseWriter, r *http.Request) {
	defer app.CloseBody(r.Body)

	name := mux.Vars(r)["name"]

	organization := organizationModel.FindByName(name)
	if organization.ID() == "" {
		response(w, http.StatusNotFound, Payload{Data: nil})
		return
	}

	response(w, http.StatusOK, Payload{Data: organization})
}
