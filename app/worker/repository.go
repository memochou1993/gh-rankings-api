package worker

import (
	"context"
	"fmt"
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/app/resource"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"strconv"
	"sync"
	"time"
)

type repositoryWorker struct {
	*Worker
	RepositoryModel *model.RepositoryModel
}

func (r *repositoryWorker) Init() {
	r.Worker.loadTimestamp(timestampRepositoryRanks)
}

func (r *repositoryWorker) Collect() error {
	logger.Info("Collecting repositories...")
	from := time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	q := model.NewRepositoriesQuery()

	return r.Travel(&from, q)
}

func (r *repositoryWorker) Travel(from *time.Time, q *model.Query) error {
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

func (r *repositoryWorker) FetchRepositories(q *model.Query, repositories *[]model.Repository) error {
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

func (r *repositoryWorker) Rank() {
	logger.Info("Executing repository rank pipelines...")
	pipelines := []*model.Pipeline{
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

	timestamp := time.Now()
	for i, p := range pipelines {
		ch <- struct{}{}
		go func(p *model.Pipeline) {
			defer wg.Done()
			RankModel.Store(r.RepositoryModel, *p, timestamp)
			<-ch
		}(p)
		if (i+1)%100 == 0 || (i+1) == len(pipelines) {
			logger.Success(fmt.Sprintf("Executed %d of %d repository rank pipelines!", i+1, len(pipelines)))
		}
	}
	wg.Wait()
	r.Worker.saveTimestamp(timestampRepositoryRanks, timestamp)
	RankModel.Delete(timestamp, model.TypeRepository)
}

func (r *repositoryWorker) fetch(q model.Query, res *model.RepositoryResponse) (err error) {
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

func (r *repositoryWorker) newSearchQuery(from time.Time) *model.SearchQuery {
	return &model.SearchQuery{
		Created: fmt.Sprintf("%s..%s", from.Format(time.RFC3339), from.AddDate(0, 0, 7).Format(time.RFC3339)),
		Fork:    "true",
		Sort:    "stars",
		Stars:   ">=100",
	}
}

func (r *repositoryWorker) newRankPipeline(field string) *model.Pipeline {
	tag := model.TypeRepository
	return &model.Pipeline{
		Pipeline: &mongo.Pipeline{
			bson.D{
				{"$project", bson.D{
					{"_id", "$_id"},
					{"image_url", "$open_graph_image_url"},
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
		Tags: []string{tag, field},
	}
}

func (r *repositoryWorker) newRankPipelinesByLanguage(field string) (pipelines []*model.Pipeline) {
	tag := model.TypeRepository
	for _, language := range resource.Languages {
		pipelines = append(pipelines, &model.Pipeline{
			Pipeline: &mongo.Pipeline{
				bson.D{
					{"$match", bson.D{
						{"primary_language.name", language.Name},
					}},
				},
				bson.D{
					{"$project", bson.D{
						{"_id", "$_id"},
						{"image_url", "$open_graph_image_url"},
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
			Tags: []string{tag, field, language.Name},
		})
	}
	return
}

func NewRepositoryWorker() *repositoryWorker {
	return &repositoryWorker{
		Worker:          NewWorker(),
		RepositoryModel: model.NewRepositoryModel(),
	}
}
