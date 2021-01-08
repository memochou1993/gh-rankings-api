package handler

import (
	"context"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/app/worker"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	repositoryModel = model.NewRepositoryModel()
)

func ListRepositories(w http.ResponseWriter, r *http.Request) {
	defer closeBody(r)

	tags := strings.Split(r.URL.Query().Get("tags"), ",")
	timestamp := worker.RepositoryWorker.Timestamp
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if page < 0 || err != nil {
		page = 1
	}
	cursor := repositoryModel.List(tags, timestamp, int(page))

	var repositories []model.Repository
	if err := cursor.All(context.Background(), &repositories); err != nil {
		log.Fatalln(err.Error())
	}

	response(w, http.StatusOK, repositories)
}
