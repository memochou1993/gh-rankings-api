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

type userWorker struct {
	*Worker
	From            time.Time
	To              time.Time
	UserModel       *model.UserModel
	SearchQuery     *model.Query
	GistQuery       *model.Query
	RepositoryQuery *model.Query
}

func (o *userWorker) Init() {
	o.Worker.loadTimestamp(timestampUserRanks)
}

func (o *userWorker) Collect() error {
	logger.Info("Collecting users...")
	o.From = time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	o.To = time.Now()

	return o.Travel()
}

func (o *userWorker) Travel() error {
	if o.From.After(o.To) {
		return nil
	}

	var users []model.User
	o.SearchQuery.SearchArguments.Query = o.buildSearchQuery()
	logger.Debug(fmt.Sprintf("User Query: %s", o.SearchQuery.SearchArguments.Query))
	if err := o.Fetch(&users); err != nil {
		return err
	}

	if res := o.UserModel.Store(users); res != nil {
		if res.ModifiedCount > 0 {
			logger.Success(fmt.Sprintf("Updated %d users!", res.ModifiedCount))
		}
		if res.UpsertedCount > 0 {
			logger.Success(fmt.Sprintf("Inserted %d users!", res.UpsertedCount))
		}
	}
	for _, user := range users {
		if err := o.Update(user); err != nil {
			return err
		}
	}
	o.From = o.From.AddDate(0, 0, 7)

	return o.Travel()
}

func (o *userWorker) Fetch(users *[]model.User) error {
	res := model.UserResponse{}
	if err := o.query(*o.SearchQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		*users = append(*users, edge.Node)
	}
	res.Data.RateLimit.Check()
	if !res.Data.Search.PageInfo.HasNextPage {
		o.SearchQuery.SearchArguments.After = ""
		return nil
	}
	o.SearchQuery.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return o.Fetch(users)
}

func (o *userWorker) Update(user model.User) error {
	o.GistQuery.Field = model.TypeUser
	o.GistQuery.OwnerArguments.Login = strconv.Quote(user.ID())
	if err := o.UpdateGists(user); err != nil {
		return err
	}

	o.RepositoryQuery.Field = model.TypeUser
	o.RepositoryQuery.OwnerArguments.Login = strconv.Quote(user.ID())
	if err := o.UpdateRepositories(user); err != nil {
		return err
	}

	return nil
}

func (o *userWorker) UpdateGists(user model.User) error {
	var gists []model.Gist
	if err := o.FetchGists(&gists); err != nil {
		return err
	}
	o.UserModel.UpdateGists(user, gists)
	logger.Success(fmt.Sprintf("Updated %d %s gists!", len(gists), model.TypeUser))
	return nil
}

func (o *userWorker) UpdateRepositories(user model.User) error {
	var repositories []model.Repository
	if err := o.FetchRepositories(&repositories); err != nil {
		return err
	}
	o.UserModel.UpdateRepositories(user, repositories)
	logger.Success(fmt.Sprintf("Updated %d %s repositories!", len(repositories), model.TypeUser))
	return nil
}

func (o *userWorker) FetchGists(gists *[]model.Gist) error {
	res := model.UserResponse{}
	if err := o.query(*o.GistQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.User.Gists.Edges {
		*gists = append(*gists, edge.Node)
	}
	res.Data.RateLimit.Check()
	if !res.Data.User.Gists.PageInfo.HasNextPage {
		o.GistQuery.GistsArguments.After = ""
		return nil
	}
	o.GistQuery.GistsArguments.After = strconv.Quote(res.Data.User.Gists.PageInfo.EndCursor)

	return o.FetchGists(gists)
}

func (o *userWorker) FetchRepositories(repositories *[]model.Repository) error {
	res := model.UserResponse{}
	if err := o.query(*o.RepositoryQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.User.Repositories.Edges {
		*repositories = append(*repositories, edge.Node)
	}
	res.Data.RateLimit.Check()
	if !res.Data.User.Repositories.PageInfo.HasNextPage {
		o.RepositoryQuery.RepositoriesArguments.After = ""
		return nil
	}
	o.RepositoryQuery.RepositoriesArguments.After = strconv.Quote(res.Data.User.Repositories.PageInfo.EndCursor)

	return o.FetchRepositories(repositories)
}

func (o *userWorker) Rank() {
	logger.Info("Executing user rank pipelines...")
	var pipelines []*model.Pipeline
	pipelines = append(pipelines, o.buildRankPipelines()...)

	ch := make(chan struct{}, 2)
	wg := sync.WaitGroup{}
	wg.Add(len(pipelines))

	now := time.Now()
	for i, p := range pipelines {
		ch <- struct{}{}
		go func(p *model.Pipeline) {
			defer wg.Done()
			RankModel.Store(o.UserModel, *p, now)
			<-ch
		}(p)
		if (i+1)%100 == 0 || (i+1) == len(pipelines) {
			logger.Success(fmt.Sprintf("Executed %d of %d user rank pipelines!", i+1, len(pipelines)))
		}
	}
	wg.Wait()
	o.Worker.saveTimestamp(timestampUserRanks, now)

	tags := []string{fmt.Sprintf("type:%s", model.TypeUser), fmt.Sprintf("type:%s", model.TypeOrganization)}
	RankModel.Delete(now, tags...)
}

func (o *userWorker) query(q model.Query, res *model.UserResponse) (err error) {
	if err := app.Fetch(context.Background(), fmt.Sprint(q), res); err != nil {
		if os.IsTimeout(err) {
			logger.Error("Retrying...")
			return o.query(q, res)
		}
		for _, err := range res.Errors {
			logger.Error(fmt.Sprintf("Error Message: %s", err.Message))
		}
		return err
	}
	for _, err := range res.Errors {
		return err
	}
	return
}

func (o *userWorker) buildSearchQuery() string {
	from := o.From.Format(time.RFC3339)
	to := o.From.AddDate(0, 0, 7).Format(time.RFC3339)
	q := model.SearchQuery{
		Created:   fmt.Sprintf("%s..%s", from, to),
		Followers: ">=100",
		Sort:      "joined-asc",
		Type:      model.TypeUser,
	}
	return strconv.Quote(util.ParseStruct(q, " "))
}

func (o *userWorker) buildRankPipelines() (pipelines []*model.Pipeline) {
	tag := fmt.Sprintf("type:%s", model.TypeUser)
	fields := []string{
		"followers",
		"gists.forks",
		"gists.stargazers",
		"repositories.forks",
		"repositories.stargazers",
		"repositories.watchers",
	}
	for _, field := range fields {
		pipelines = append(pipelines, o.buildRankPipeline(field, tag))
		pipelines = append(pipelines, o.buildRankPipelinesByLocation(field, tag)...)
	}
	pipelines = append(pipelines, o.buildRepositoryRankPipelinesByLanguage("forks", tag)...)
	pipelines = append(pipelines, o.buildRepositoryRankPipelinesByLanguage("stargazers", tag)...)
	pipelines = append(pipelines, o.buildRepositoryRankPipelinesByLanguage("watchers", tag)...)
	return
}

func (o *userWorker) buildRankPipeline(field string, tags ...string) *model.Pipeline {
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
					{"image_url", "$avatar_url"},
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
		Tags: append(tags, fmt.Sprintf("field:%s", field)),
	}
}

func (o *userWorker) buildRankPipelinesByLocation(field string, tags ...string) (pipelines []*model.Pipeline) {
	for _, location := range resource.Locations {
		pipelines = append(pipelines, o.buildRankPipeline(field, append(tags, fmt.Sprintf("location:%s", location.Name))...))
		for _, city := range location.Cities {
			pipelines = append(pipelines, o.buildRankPipeline(field, append(tags, fmt.Sprintf("location:%s, %s", city.Name, location.Name))...))
		}
	}
	return
}

func (o *userWorker) buildRepositoryRankPipelinesByLanguage(field string, tags ...string) (pipelines []*model.Pipeline) {
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
						{"image_url", bson.D{
							{"$first", "$avatar_url"},
						}},
						{"total_count", bson.D{
							{"$sum", fmt.Sprintf("$repositories.%s.total_count", field)},
						}},
					}},
				},
				bson.D{
					{"$sort", bson.D{
						{"total_count", -1},
					}},
				},
			},
			Tags: append(tags, fmt.Sprintf("field:repositories.%s", field), fmt.Sprintf("language:%s", language.Name)),
		})
	}
	return
}

func NewUserWorker() *userWorker {
	return &userWorker{
		Worker:          NewWorker(),
		UserModel:       model.NewUserModel(),
		SearchQuery:     model.NewOwnerQuery(),
		GistQuery:       model.NewOwnerGistQuery(),
		RepositoryQuery: model.NewOwnerRepositoryQuery(),
	}
}
