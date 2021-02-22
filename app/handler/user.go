package handler

import (
	"github.com/gorilla/mux"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/model"
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

	users := userModel.List(req)

	response(w, http.StatusOK, Payload{Data: users})
}

func ShowUser(w http.ResponseWriter, r *http.Request) {
	defer app.CloseBody(r.Body)

	name := mux.Vars(r)["name"]

	user := userModel.FindByName(name)
	if user.ID() == "" {
		response(w, http.StatusNotFound, Payload{Data: nil})
		return
	}

	response(w, http.StatusOK, Payload{Data: user})
}
