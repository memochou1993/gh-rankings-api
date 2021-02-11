package worker

import (
	"context"
	"errors"
	"fmt"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/pipeline"
	"github.com/memochou1993/gh-rankings/app/response"
	"github.com/memochou1993/gh-rankings/logger"
	"github.com/memochou1993/gh-rankings/util"
	"os"
	"strconv"
	"sync"
	"time"
)

type repositoryWorker struct {
	*Worker
	From            time.Time
	To              time.Time
	RepositoryModel *model.RepositoryModel
	SearchQuery     *model.Query
}

func (r *repositoryWorker) Collect() error {
	logger.Info("Collecting repositories...")
	r.From = time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	r.To = time.Now()

	if r.Worker.Timestamp.IsZero() {
		if repository := r.RepositoryModel.FindLast(); repository.ID() != "" {
			r.From = repository.CreatedAt.AddDate(0, 0, -7).Truncate(24 * time.Hour)
		}
	}

	return r.Travel()
}

func (r *repositoryWorker) Travel() error {
	if r.From.After(r.To) {
		return nil
	}

	var repositories []model.Repository
	r.SearchQuery.SearchArguments.Query = r.buildSearchQuery()
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

func (r *repositoryWorker) Fetch(repositories *[]model.Repository) error {
	res := response.Repository{}
	if err := r.query(*r.SearchQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		*repositories = append(*repositories, edge.Node)
	}
	res.Data.RateLimit.Break(collecting)
	if !res.Data.Search.PageInfo.HasNextPage {
		r.SearchQuery.SearchArguments.After = ""
		return nil
	}
	r.SearchQuery.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return r.Fetch(repositories)
}

func (r *repositoryWorker) Rank() {
	logger.Info("Executing repository rank pipelines...")
	pipelines := r.buildRankPipelines()

	ch := make(chan struct{}, 2)
	wg := sync.WaitGroup{}
	wg.Add(len(pipelines))

	timestamp := time.Now()
	for i, p := range pipelines {
		ch <- struct{}{}
		go func(p *pipeline.Pipeline) {
			defer wg.Done()
			RankModel.Store(r.RepositoryModel, *p, timestamp)
			<-ch
		}(p)
		if (i+1)%10 == 0 || (i+1) == len(pipelines) {
			logger.Success(fmt.Sprintf("Executed %d of %d repository rank pipelines!", i+1, len(pipelines)))
		}
	}
	wg.Wait()
	r.Worker.seal(TimestampRepositoryRanks, timestamp)

	RankModel.Delete(timestamp, model.TypeRepository)
}

func (r *repositoryWorker) query(q model.Query, res *response.Repository) (err error) {
	if err = app.Fetch(context.Background(), fmt.Sprint(q), res); err != nil {
		if !os.IsTimeout(err) {
			return err
		}
	}
	if res.Message != "" {
		err = errors.New(res.Message)
	}
	for _, err = range res.Errors {
		break
	}
	if err != nil {
		logger.Error(err.Error())
		logger.Warning("Retrying...")
		time.Sleep(10 * time.Second)
		return r.query(q, res)
	}
	return
}

func (r *repositoryWorker) buildSearchQuery() string {
	from := r.From.Format(time.RFC3339)
	to := r.From.AddDate(0, 0, 7).Format(time.RFC3339)
	q := model.SearchQuery{
		Created: fmt.Sprintf("%s..%s", from, to),
		Fork:    "true",
		Sort:    "stars",
		Stars:   ">=100",
	}
	return strconv.Quote(util.ParseStruct(q, " "))
}

func (r *repositoryWorker) buildRankPipelines() (pipelines []*pipeline.Pipeline) {
	rankType := model.TypeRepository
	fields := []string{
		"forks",
		"stargazers",
		"watchers",
	}
	for _, field := range fields {
		pipelines = append(pipelines, pipeline.RankByField(rankType, field))
		pipelines = append(pipelines, pipeline.RankRepositoryByLanguage(rankType, field)...)
	}
	return
}

func NewRepositoryWorker() *repositoryWorker {
	return &repositoryWorker{
		Worker:          NewWorker(),
		RepositoryModel: model.NewRepositoryModel(),
		SearchQuery:     model.NewRepositoryQuery(),
	}
}
