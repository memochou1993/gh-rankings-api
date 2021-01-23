package handler

import (
	"context"
	"github.com/memochou1993/github-rankings/app/handler/request"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/app/worker"
	"log"
	"net/http"
)

func ListOwners(w http.ResponseWriter, r *http.Request) {
	defer closeBody(r)

	req := request.NewOwnerRequest(r)
	req.CreatedAt = worker.OwnerWorker.Timestamp

	var owners []model.OwnerRank
	cursor := model.NewOwnerRankModel().List(req)
	if err := cursor.All(context.Background(), &owners); err != nil {
		log.Fatalln(err.Error())
	}

	response(w, http.StatusOK, owners)
}
