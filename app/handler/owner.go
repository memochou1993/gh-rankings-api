package handler

import (
	"github.com/gorilla/mux"
	"github.com/memochou1993/github-rankings/app/model"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

var ownerModel *model.OwnerModel

func init() {
	ownerModel = model.NewOwnerModel()
}

func ListOwners(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// TODO
}

func ShowOwner(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	login := mux.Vars(r)["login"]
	res := ownerModel.Find(login)
	if res.Err() == mongo.ErrNoDocuments {
		response(w, http.StatusNotFound, nil)
		return
	}
	if res.Err() != nil {
		response(w, http.StatusInternalServerError, nil)
		return
	}

	var owner model.Owner
	if err := res.Decode(&owner); err != nil {
		response(w, http.StatusInternalServerError, nil)
		return
	}

	response(w, http.StatusOK, owner)
}
