package handler

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/app/worker"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	ownerModel = model.NewOwnerModel()
)

func ListOwners(w http.ResponseWriter, r *http.Request) {
	defer closeBody(r)

	tags := strings.Split(r.URL.Query().Get("tags"), ",")
	timestamp := worker.OwnerWorker.Timestamp
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if page < 0 || err != nil {
		page = 1
	}

	var owners []model.OwnerRank
	if timestamp == nil {
		response(w, http.StatusOK, owners)
		return
	}
	cursor := model.NewRankModel().List(model.NewOwnerRankModel(), tags, *timestamp, int(page))
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
