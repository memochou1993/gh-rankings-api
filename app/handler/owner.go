package handler

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/memochou1993/github-rankings/app/handler/request"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/app/worker"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
)

var (
	ownerModel = model.NewOwnerModel()
)

func ListOwners(w http.ResponseWriter, r *http.Request) {
	defer closeBody(r)

	req := request.NewOwnerRequest(r)
	req.CreatedAt = worker.OwnerWorker.Timestamp

	var owners []model.OwnerRank
	if req.CreatedAt == nil {
		response(w, http.StatusAccepted, owners)
		return
	}
	cursor := model.NewOwnerRankModel().List(req)
	if err := cursor.All(context.Background(), &owners); err != nil {
		log.Fatalln(err.Error())
	}

	response(w, http.StatusOK, owners)
}

func ShowOwner(w http.ResponseWriter, r *http.Request) {
	defer closeBody(r)

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
