package worker

import (
	"context"
	"errors"
	"fmt"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/pipeline"
	"github.com/memochou1993/gh-rankings/app/query"
	"github.com/memochou1993/gh-rankings/app/response"
	"github.com/memochou1993/gh-rankings/logger"
	"os"
	"strconv"
	"time"
)

type Repository struct {
	*Worker
	From            time.Time
	To              time.Time
	RepositoryModel *model.RepositoryModel
	RankModel       *model.RankModel
	SearchQuery     *query.Query
}

func (r *Repository) Init() {
	r.Worker.load(TimestampRepository)
}

func (r *Repository) Collect() error {
	logger.Info("Collecting repositories...")
	r.From = time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	r.To = time.Now()

	if r.Worker.Timestamp.IsZero() {
		last := model.Repository{}
		if r.RepositoryModel.Model.Last(&last); last.ID() != "" {
			r.From = last.CreatedAt.AddDate(0, 0, -7).Truncate(24 * time.Hour)
		}
	}

	return r.Travel()
}

func (r *Repository) Travel() error {
	if r.From.After(r.To) {
		return nil
	}

	var repositories []model.Repository
	r.SearchQuery.SearchArguments.SetQuery(query.SearchRepositories(r.From, r.From.AddDate(0, 0, 7)))
	logger.Debug(fmt.Sprintf("Repository Query: %s", r.SearchQuery.SearchArguments.Query))
	if err := r.Fetch(&repositories); err != nil {
		return err
	}
	if res := r.RepositoryModel.Store(repositories); res != nil {
		if res.ModifiedCount > 0 {
			logger.Success(fmt.Sprintf("Updated %d repositories!", res.ModifiedCount))
		}
		if res.UpsertedCount > 0 {
			logger.Success(fmt.Sprintf("Inserted %d repositories!", res.UpsertedCount))
		}
	}
	r.From = r.From.AddDate(0, 0, 7)

	return r.Travel()
}

func (r *Repository) Fetch(repositories *[]model.Repository) error {
	res := response.Repository{}
	if err := r.query(*r.SearchQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		*repositories = append(*repositories, edge.Node)
	}
	res.Data.RateLimit.Throttle(collecting)
	if !res.Data.Search.PageInfo.HasNextPage {
		r.SearchQuery.SearchArguments.After = ""
		return nil
	}
	r.SearchQuery.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return r.Fetch(repositories)
}

func (r *Repository) Rank() {
	logger.Info("Executing repository rank pipelines...")
	pipelines := pipeline.Repository()
	timestamp := time.Now()
	for i, p := range pipelines {
		r.RankModel.Store(r.RepositoryModel, *p, timestamp)
		if (i+1)%10 == 0 || (i+1) == len(pipelines) {
			logger.Success(fmt.Sprintf("Executed %d of %d repository rank pipelines!", i+1, len(pipelines)))
		}
	}
	r.Worker.save(TimestampRepository, timestamp)
	r.RankModel.Delete(timestamp, model.TypeRepository)
}

func (r *Repository) query(q query.Query, res *response.Repository) (err error) {
	if err = app.Fetch(context.Background(), fmt.Sprint(q), res); err != nil {
		if !os.IsTimeout(err) {
			return err
		}
	}
	if res.Message != "" {
		err = errors.New(res.Message)
		res.Message = ""
	}
	for _, err = range res.Errors {
		return err
	}
	if err != nil {
		logger.Error(err.Error())
		logger.Warning("Retrying...")
		time.Sleep(10 * time.Second)
		return r.query(q, res)
	}
	return
}

func NewRepositoryWorker() *Repository {
	return &Repository{
		Worker:          &Worker{},
		RepositoryModel: model.NewRepositoryModel(),
		RankModel:       model.NewRankModel(),
		SearchQuery:     query.Repositories(),
	}
}
