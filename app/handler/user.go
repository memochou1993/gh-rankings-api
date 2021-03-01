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
	userModel = model.NewUserModel()
)

func ListUsers(w http.ResponseWriter, r *http.Request) {
	defer app.CloseBody(r.Body)

	req, err := request.NewUserRequest(r)
	if err != nil {
		response(w, http.StatusUnprocessableEntity, Payload{Error: err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("%s:%s", app.TypeUser, fmt.Sprint(req))
	items, found := app.Cache.Get(cacheKey)
	if !found {
		items = userModel.List(req)
		app.Cache.Set(cacheKey, &items, cache.DefaultExpiration)
	}

	response(w, http.StatusOK, Payload{Data: items})
}

func ShowUser(w http.ResponseWriter, r *http.Request) {
	defer app.CloseBody(r.Body)

	id := mux.Vars(r)["login"]

	cacheKey := fmt.Sprintf("%s:%s", app.TypeUser, id)
	item, found := app.Cache.Get(cacheKey)
	if !found {
		user := userModel.FindByID(id)
		if user.ID() == "" {
			response(w, http.StatusNotFound, Payload{Data: nil})
			return
		}
		app.Cache.Set(cacheKey, &user, cache.DefaultExpiration)
		item = user
	}

	response(w, http.StatusOK, Payload{Data: item})
}
