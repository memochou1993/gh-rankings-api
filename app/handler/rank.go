package handler

import (
	"fmt"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/worker"
	"github.com/patrickmn/go-cache"
	"net/http"
)

var (
	rankModel = model.NewRankModel()
)

func ListRanks(w http.ResponseWriter, r *http.Request) {
	defer app.CloseBody(r.Body)

	req, err := request.NewRankRequest(r)
	if err != nil {
		response(w, http.StatusUnprocessableEntity, Payload{Error: err.Error()})
		return
	}

	switch req.Type {
	case app.TypeUser:
		req.Timestamps = append(req.Timestamps, worker.UserWorker.Timestamp)
	case app.TypeOrganization:
		req.Timestamps = append(req.Timestamps, worker.OrganizationWorker.Timestamp)
	case app.TypeRepository:
		req.Timestamps = append(req.Timestamps, worker.RepositoryWorker.Timestamp)
	default:
		req.Timestamps = append(req.Timestamps, worker.UserWorker.Timestamp)
		req.Timestamps = append(req.Timestamps, worker.OrganizationWorker.Timestamp)
		req.Timestamps = append(req.Timestamps, worker.RepositoryWorker.Timestamp)
	}

	cacheKey := fmt.Sprint(req)
	ranks, found := app.Cache.Get(cacheKey)
	if !found {
		ranks = rankModel.List(req)
		app.Cache.Set(cacheKey, &ranks, cache.DefaultExpiration)
	}

	response(w, http.StatusOK, Payload{Data: ranks})
}
