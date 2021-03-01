package handler

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/patrickmn/go-cache"
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

	cacheKey := fmt.Sprintf("%s:%s", app.TypeOrganization, fmt.Sprint(req))
	items, found := app.Cache.Get(cacheKey)
	if !found {
		items = organizationModel.List(req)
		app.Cache.Set(cacheKey, &items, cache.DefaultExpiration)
	}

	response(w, http.StatusOK, Payload{Data: items})
}

func ShowOrganization(w http.ResponseWriter, r *http.Request) {
	defer app.CloseBody(r.Body)

	id := mux.Vars(r)["login"]

	cacheKey := fmt.Sprintf("%s:%s", app.TypeOrganization, id)
	item, found := app.Cache.Get(cacheKey)
	if !found {
		organization := organizationModel.FindByID(id)
		if organization.ID() == "" {
			response(w, http.StatusNotFound, Payload{Data: nil})
			return
		}
		app.Cache.Set(cacheKey, &organization, cache.DefaultExpiration)
		item = organization
	}

	response(w, http.StatusOK, Payload{Data: item})
}
