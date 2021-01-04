package handler

import (
	"github.com/gorilla/mux"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
)

var ownerModel *model.OwnerModel

func init() {
	ownerModel = model.NewOwnerModel()
}

func ShowOwner(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	login := mux.Vars(r)["login"]
	projection := bson.D{
		{"gists", 0},
		{"repositories", 0},
	}

	opts := options.FindOne().SetProjection(projection)
	res := database.FindOne(ownerModel.Name(), bson.D{{"_id", login}}, opts)
	if res.Err() == mongo.ErrNoDocuments {
		response(w, http.StatusNotFound, nil)
		return
	}

	var owner model.Owner
	res.Decode(&owner)

	response(w, http.StatusOK, owner)
}
