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
	repositoryModel = model.NewRepositoryModel()
)

func ListRepositories(w http.ResponseWriter, r *http.Request) {
	defer app.CloseBody(r.Body)

	req, err := request.NewRepositoryRequest(r)
	if err != nil {
		response(w, http.StatusUnprocessableEntity, Payload{Error: err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("%s:%s", app.TypeRepository, fmt.Sprint(req))
	items, found := app.Cache.Get(cacheKey)
	if !found {
		items = repositoryModel.List(req)
		app.Cache.Set(cacheKey, &items, cache.DefaultExpiration)
	}

	response(w, http.StatusOK, Payload{Data: items})
}

func ShowRepository(w http.ResponseWriter, r *http.Request) {
	defer app.CloseBody(r.Body)

	id := fmt.Sprintf("%s/%s", mux.Vars(r)["owner"], mux.Vars(r)["name"])

	cacheKey := fmt.Sprintf("%s:%s", app.TypeRepository, id)
	item, found := app.Cache.Get(cacheKey)
	if !found {
		repository := repositoryModel.FindByID(id)
		if repository.ID() == "" {
			response(w, http.StatusNotFound, Payload{Data: nil})
			return
		}
		app.Cache.Set(cacheKey, &repository, cache.DefaultExpiration)
		item = repository
	}

	response(w, http.StatusOK, Payload{Data: item})
}
