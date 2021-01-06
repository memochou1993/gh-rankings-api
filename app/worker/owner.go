package worker

import (
	"context"
	"fmt"
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/app/resource"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type OwnerWorker struct {
	OwnerModel *model.OwnerModel
	Timestamp  time.Time
}

func (o *OwnerWorker) Init() {
	logger.Info("Initializing owner collection...")
	o.OwnerModel.CreateIndexes()
	logger.Success("Owner collection initialized!")
}

func (o *OwnerWorker) Work() {
	go func() {
		t := time.NewTicker(10 * time.Second)
		for ; true; <-t.C {
			if err := o.Collect(); err != nil {
				logger.Error(err.Error())
			}
		}
	}()

	go func() {
		t := time.NewTicker(10 * time.Minute)
		for ; true; <-t.C {
			if err := o.Update(); err != nil {
				logger.Error(err.Error())
			}
		}
	}()

	go func() {
		t := time.NewTicker(24 * time.Hour)
		for ; true; <-t.C {
			o.Rank()
		}
	}()
}

func (o *OwnerWorker) Collect() error {
	logger.Info("Collecting owners...")
	from := time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	q := model.NewOwnersQuery()

	return o.Travel(&from, q)
}

func (o *OwnerWorker) Travel(from *time.Time, q *model.Query) error {
	to := time.Now()
	if from.After(to) {
		return nil
	}

	q.SearchArguments.Query = strconv.Quote(util.ParseStruct(o.newSearchQuery(*from), " "))

	var owners []model.Owner
	if err := o.FetchOwners(q, &owners); err != nil {
		return err
	}
	if res := o.OwnerModel.Store(owners); res != nil {
		if res.ModifiedCount > 0 {
			logger.Success(fmt.Sprintf("Updated %d owners!", res.ModifiedCount))
		}
		if res.UpsertedCount > 0 {
			logger.Success(fmt.Sprintf("Inserted %d owners!", res.UpsertedCount))
		}
	}
	*from = from.AddDate(0, 0, 7)

	return o.Travel(from, q)
}

func (o *OwnerWorker) FetchOwners(q *model.Query, owners *[]model.Owner) error {
	res := model.OwnerResponse{}
	if err := o.fetch(*q, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		*owners = append(*owners, edge.Node)
	}
	res.Data.RateLimit.Break()
	if !res.Data.Search.PageInfo.HasNextPage {
		q.SearchArguments.After = ""
		return nil
	}
	q.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return o.FetchOwners(q, owners)
}

func (o *OwnerWorker) Update() error {
	logger.Info("Updating owner gists...")
	logger.Info("Updating owner repositories...")
	ctx := context.Background()
	stop := false
	limit := 100
	gistsQuery := model.NewOwnerGistsQuery()
	repositoriesQuery := model.NewOwnerRepositoriesQuery()
	for i := 0; !stop; i++ {
		err := func() error {
			cursor := database.All(ctx, o.OwnerModel.Name(), i*limit, limit)
			defer database.CloseCursor(ctx, cursor)
			if cursor.RemainingBatchLength() == 0 {
				stop = true
			}
			for cursor.Next(context.Background()) {
				owner := model.Owner{}
				if err := cursor.Decode(&owner); err != nil {
					log.Fatalln(err.Error())
				}
				if err := o.UpdateGists(owner, gistsQuery); err != nil {
					return err
				}
				if err := o.UpdateRepositories(owner, repositoriesQuery); err != nil {
					return err
				}
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *OwnerWorker) UpdateGists(owner model.Owner, q *model.Query) error {
	if !owner.IsUser() {
		return nil
	}
	var gists []model.Gist
	q.OwnerArguments.Login = strconv.Quote(owner.ID())
	if err := o.FetchGists(q, &gists); err != nil {
		return err
	}
	o.OwnerModel.UpdateGists(owner, gists)
	logger.Success(fmt.Sprintf("Updated %d %s gists!", len(gists), owner.Type()))

	return nil
}

func (o *OwnerWorker) UpdateRepositories(owner model.Owner, q *model.Query) error {
	var repositories []model.Repository
	q.Field = owner.Type()
	q.OwnerArguments.Login = strconv.Quote(owner.ID())
	if err := o.FetchRepositories(q, &repositories); err != nil {
		return err
	}
	o.OwnerModel.UpdateRepositories(owner, repositories)
	logger.Success(fmt.Sprintf("Updated %d %s repositories!", len(repositories), owner.Type()))

	return nil
}

func (o *OwnerWorker) FetchGists(q *model.Query, gists *[]model.Gist) error {
	res := model.OwnerResponse{}
	if err := o.fetch(*q, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Owner.Gists.Edges {
		*gists = append(*gists, edge.Node)
	}
	res.Data.RateLimit.Break()
	if !res.Data.Owner.Gists.PageInfo.HasNextPage {
		q.GistsArguments.After = ""
		return nil
	}
	q.GistsArguments.After = strconv.Quote(res.Data.Owner.Gists.PageInfo.EndCursor)

	return o.FetchGists(q, gists)
}

func (o *OwnerWorker) FetchRepositories(q *model.Query, repositories *[]model.Repository) error {
	res := model.OwnerResponse{}
	if err := o.fetch(*q, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Owner.Repositories.Edges {
		*repositories = append(*repositories, edge.Node)
	}
	res.Data.RateLimit.Break()
	if !res.Data.Owner.Repositories.PageInfo.HasNextPage {
		q.RepositoriesArguments.After = ""
		return nil
	}
	q.RepositoriesArguments.After = strconv.Quote(res.Data.Owner.Repositories.PageInfo.EndCursor)

	return o.FetchRepositories(q, repositories)
}

func (o *OwnerWorker) Rank() {
	logger.Info("Executing owner rank pipelines...")
	var pipelines []*model.Pipeline
	pipelines = append(pipelines, o.newUserRankPipelines()...)
	pipelines = append(pipelines, o.newOrganizationRankPipelines()...)

	ch := make(chan struct{}, 4)
	wg := sync.WaitGroup{}
	wg.Add(len(pipelines))
	timestamp := time.Now()
	for _, p := range pipelines {
		ch <- struct{}{}
		go func(p *model.Pipeline) {
			defer wg.Done()
			model.PushRanks(o.OwnerModel, timestamp, *p)
			<-ch
		}(p)
	}
	wg.Wait()
	o.Timestamp = timestamp
	model.PullRanks(o.OwnerModel, timestamp)
	logger.Success(fmt.Sprintf("Executed %d owner rank pipelines!", len(pipelines)))
}

func (o *OwnerWorker) fetch(q model.Query, res *model.OwnerResponse) (err error) {
	if err := app.Fetch(context.Background(), fmt.Sprint(q), res); err != nil {
		if os.IsTimeout(err) {
			logger.Error("Retrying...")
			return o.fetch(q, res)
		}
		return err
	}
	for _, err := range res.Errors {
		return err
	}
	return
}

func (o *OwnerWorker) newSearchQuery(from time.Time) *model.SearchQuery {
	return &model.SearchQuery{
		Created: fmt.Sprintf("%s..%s", from.Format(time.RFC3339), from.AddDate(0, 0, 7).Format(time.RFC3339)),
		Repos:   ">=5",
		Sort:    "joined-asc",
	}
}

func (o *OwnerWorker) newUserRankPipelines() (pipelines []*model.Pipeline) {
	tag := model.TypeUser
	fields := []string{
		"followers",
		"gists.forks",
		"gists.stargazers",
		"repositories.forks",
		"repositories.stargazers",
		"repositories.watchers",
	}
	for _, field := range fields {
		pipelines = append(pipelines, o.newRankPipeline(field, tag))
		pipelines = append(pipelines, o.newRankPipelinesByLocation(field, tag)...)
	}
	pipelines = append(pipelines, o.newRepositoryRankPipelinesByLanguage("forks", tag)...)
	pipelines = append(pipelines, o.newRepositoryRankPipelinesByLanguage("stargazers", tag)...)
	pipelines = append(pipelines, o.newRepositoryRankPipelinesByLanguage("watchers", tag)...)
	return
}

func (o *OwnerWorker) newOrganizationRankPipelines() (pipelines []*model.Pipeline) {
	tag := model.TypeOrganization
	fields := []string{
		"repositories.forks",
		"repositories.stargazers",
		"repositories.watchers",
	}
	for _, field := range fields {
		pipelines = append(pipelines, o.newRankPipeline(field, tag))
		pipelines = append(pipelines, o.newRankPipelinesByLocation(field, tag)...)
	}
	pipelines = append(pipelines, o.newRepositoryRankPipelinesByLanguage("forks", tag)...)
	pipelines = append(pipelines, o.newRepositoryRankPipelinesByLanguage("stargazers", tag)...)
	pipelines = append(pipelines, o.newRepositoryRankPipelinesByLanguage("watchers", tag)...)
	return
}

func (o *OwnerWorker) newRankPipeline(field string, tags ...string) *model.Pipeline {
	return &model.Pipeline{
		Pipeline: &mongo.Pipeline{
			bson.D{
				{"$match", bson.D{
					{"tags", bson.D{
						{"$all", tags},
					}},
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
		Tags: append(tags, field),
	}
}

func (o *OwnerWorker) newRankPipelinesByLocation(field string, tags ...string) (pipelines []*model.Pipeline) {
	for _, location := range resource.Locations {
		pipelines = append(pipelines, o.newRankPipeline(field, append(tags, location.Name)...))
		for _, city := range location.Cities {
			pipelines = append(pipelines, o.newRankPipeline(field, append(tags, city.Name)...))
		}
	}
	return
}

func (o *OwnerWorker) newRepositoryRankPipelinesByLanguage(field string, tags ...string) (pipelines []*model.Pipeline) {
	for _, language := range resource.Languages {
		pipelines = append(pipelines, &model.Pipeline{
			Pipeline: &mongo.Pipeline{
				bson.D{
					{"$match", bson.D{
						{"tags", bson.D{
							{"$all", tags},
						}},
					}},
				},
				bson.D{
					{"$unwind", "$repositories"},
				},
				bson.D{
					{"$match", bson.D{
						{"repositories.primary_language.name", language.Name},
					}},
				},
				bson.D{
					{"$group", bson.D{
						{"_id", "$_id"},
						{"total_count", bson.D{
							{"$sum", fmt.Sprintf("$repositories.%s.total_count", field)},
						}},
					}},
				},
				bson.D{
					{"$match", bson.D{
						{"total_count", bson.D{
							{"$gt", 0},
						}},
					}},
				},
				bson.D{
					{"$sort", bson.D{
						{"total_count", -1},
					}},
				},
			},
			Tags: append(tags, fmt.Sprintf("repositories.%s", field), language.Name),
		})
	}
	return
}

func NewOwnerWorker() *OwnerWorker {
	return &OwnerWorker{
		OwnerModel: model.NewOwnerModel(),
	}
}
