package worker

import (
	"context"
	"fmt"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/resource"
	"github.com/memochou1993/gh-rankings/logger"
	"github.com/memochou1993/gh-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
	RepositoryQuery *model.Query
}

func (r *repositoryWorker) Init() {
	r.Worker.loadTimestamp(timestampRepositoryRanks)
}

func (r *repositoryWorker) Collect() error {
	logger.Info("Collecting repositories...")
	r.From = time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	r.To = time.Now()

	return r.Travel()
}

func (r *repositoryWorker) Travel() error {
	if r.From.After(r.To) {
		return nil
	}

	repositories := map[string]model.Repository{}

	r.RepositoryQuery.SearchArguments.Query = r.buildSearchQuery()
	logger.Debug(fmt.Sprintf("Repository Query: %s", r.RepositoryQuery.SearchArguments.Query))

	if err := r.FetchRepositories(repositories); err != nil {
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

func (r *repositoryWorker) FetchRepositories(repositories map[string]model.Repository) error {
	res := model.RepositoryResponse{}
	if err := r.fetch(*r.RepositoryQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		repositories[edge.Node.NameWithOwner] = edge.Node
	}
	res.Data.RateLimit.Check()
	if !res.Data.Search.PageInfo.HasNextPage {
		r.RepositoryQuery.SearchArguments.After = ""
		return nil
	}
	r.RepositoryQuery.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return r.FetchRepositories(repositories)
}

func (r *repositoryWorker) Rank() {
	logger.Info("Executing repository rank pipelines...")
	pipelines := []*model.Pipeline{
		r.buildRankPipeline("forks"),
		r.buildRankPipeline("stargazers"),
		r.buildRankPipeline("watchers"),
	}
	pipelines = append(pipelines, r.buildRankPipelinesByLanguage("forks")...)
	pipelines = append(pipelines, r.buildRankPipelinesByLanguage("stargazers")...)
	pipelines = append(pipelines, r.buildRankPipelinesByLanguage("watchers")...)

	ch := make(chan struct{}, 2)
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

	tag := fmt.Sprintf("type:%s", model.TypeRepository)
	RankModel.Delete(timestamp, tag)
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

func (r *repositoryWorker) buildRankPipeline(field string) *model.Pipeline {
	tag := fmt.Sprintf("type:%s", model.TypeRepository)
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
		Tags: []string{tag, fmt.Sprintf("field:%s", field)},
	}
}

func (r *repositoryWorker) buildRankPipelinesByLanguage(field string) (pipelines []*model.Pipeline) {
	tag := fmt.Sprintf("type:%s", model.TypeRepository)
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
			Tags: []string{tag, fmt.Sprintf("field:%s", field), fmt.Sprintf("language:%s", language.Name)},
		})
	}
	return
}

func NewRepositoryWorker() *repositoryWorker {
	return &repositoryWorker{
		Worker:          NewWorker(),
		RepositoryModel: model.NewRepositoryModel(),
		RepositoryQuery: model.NewRepositoryQuery(),
	}
}
