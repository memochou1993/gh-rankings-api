package worker

import (
	"context"
	"fmt"
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"strconv"
	"sync"
	"time"
)

type RepositoryWorker struct {
	RepositoryModel *model.RepositoryModel
	UpdatedAt       time.Time
}

func NewRepositoryWorker() *RepositoryWorker {
	return &RepositoryWorker{
		RepositoryModel: model.NewRepositoryModel(),
	}
}

func (r *RepositoryWorker) Init() {
	logger.Info("Initializing repository collection...")
	r.RepositoryModel.CreateIndexes()
	logger.Success("Repository collection initialized!")
}

func (r *RepositoryWorker) Collect() error {
	logger.Info("Collecting repositories...")
	from := time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	q := model.NewRepositoriesQuery()

	return r.Travel(&from, q)
}

func (r *RepositoryWorker) Travel(from *time.Time, q *model.Query) error {
	to := time.Now()
	if from.After(to) {
		return nil
	}

	q.SearchArguments.Query = strconv.Quote(util.ParseStruct(r.newSearchQuery(*from), " "))

	var repositories []model.Repository
	if err := r.FetchRepositories(q, &repositories); err != nil {
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
	*from = from.AddDate(0, 0, 7)

	return r.Travel(from, q)
}

func (r *RepositoryWorker) FetchRepositories(q *model.Query, repositories *[]model.Repository) error {
	res := model.RepositoryResponse{}
	if err := r.fetch(*q, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		*repositories = append(*repositories, edge.Node)
	}
	res.Data.RateLimit.Break()
	if !res.Data.Search.PageInfo.HasNextPage {
		q.SearchArguments.After = ""
		return nil
	}
	q.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return r.FetchRepositories(q, repositories)
}

func (r *RepositoryWorker) Rank() {
	logger.Info("Executing repository rank pipelines...")
	pipelines := []*model.RankPipeline{
		r.newRankPipeline("forks"),
		r.newRankPipeline("stargazers"),
		r.newRankPipeline("watchers"),
	}
	pipelines = append(pipelines, r.newRankPipelinesByLanguage("forks")...)
	pipelines = append(pipelines, r.newRankPipelinesByLanguage("stargazers")...)
	pipelines = append(pipelines, r.newRankPipelinesByLanguage("watchers")...)

	ch := make(chan struct{}, 4)
	wg := sync.WaitGroup{}
	wg.Add(len(pipelines))
	updatedAt := time.Now()
	for _, pipeline := range pipelines {
		ch <- struct{}{}
		go func(pipeline *model.RankPipeline) {
			defer wg.Done()
			model.PushRanks(r.RepositoryModel, updatedAt, *pipeline)
			<-ch
		}(pipeline)
	}
	wg.Wait()
	r.UpdatedAt = updatedAt
	model.PullRanks(r.RepositoryModel, updatedAt)
	logger.Success(fmt.Sprintf("Executed %d repository rank pipelines!", len(pipelines)))
}

func (r *RepositoryWorker) fetch(q model.Query, res *model.RepositoryResponse) (err error) {
	if err := app.Fetch(context.Background(), fmt.Sprint(q), res); err != nil {
		if os.IsTimeout(err) {
			logger.Error("Retrying...")
			return r.fetch(q, res)
		}
		return err
	}
	for _, err := range res.Errors {
		return err
	}
	return
}

func (r *RepositoryWorker) newSearchQuery(from time.Time) *model.SearchQuery {
	return &model.SearchQuery{
		Created: fmt.Sprintf("%s..%s", from.Format(time.RFC3339), from.AddDate(0, 0, 7).Format(time.RFC3339)),
		Fork:    "true",
		Sort:    "stars",
		Stars:   ">=100",
	}
}

func (r *RepositoryWorker) newRankPipeline(field string) *model.RankPipeline {
	return &model.RankPipeline{
		Pipeline: &mongo.Pipeline{
			bson.D{
				{"$project", bson.D{
					{"_id", "$_id"},
					{"total_count", bson.D{
						{"$sum", fmt.Sprintf("$%s.total_count", field)},
					}},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"total_count", -1},
				}},
			},
		},
		Tags: []string{model.TypeRepository, field},
	}
}

func (r *RepositoryWorker) newRankPipelinesByLanguage(field string) (pipelines []*model.RankPipeline) {
	for _, language := range languages {
		pipelines = append(pipelines, &model.RankPipeline{
			Pipeline: &mongo.Pipeline{
				bson.D{
					{"$match", bson.D{
						{"primary_language.name", language.Name},
					}},
				},
				bson.D{
					{"$project", bson.D{
						{"_id", "$_id"},
						{"total_count", bson.D{
							{"$sum", fmt.Sprintf("$%s.total_count", field)},
						}},
					}},
				},
				bson.D{
					{"$sort", bson.D{
						{"total_count", -1},
					}},
				},
			},
			Tags: []string{model.TypeRepository, field, language.Name},
		})
	}
	return
}
