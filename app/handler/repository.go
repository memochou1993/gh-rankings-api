package handler

import (
	"context"
	"github.com/memochou1993/github-rankings/app/handler/request"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/app/worker"
	"log"
	"net/http"
)

func ListRepositories(w http.ResponseWriter, r *http.Request) {
	defer closeBody(r)

	req := request.NewRepositoryRequest(r)
	req.CreatedAt = worker.RepositoryWorker.Timestamp

	var repositories []model.RepositoryRank
	cursor := model.NewRepositoryRankModel().List(req)
	if err := cursor.All(context.Background(), &repositories); err != nil {
		log.Fatalln(err.Error())
	}

	response(w, http.StatusOK, repositories)
}
