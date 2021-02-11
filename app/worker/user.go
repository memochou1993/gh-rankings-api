package worker

import (
	"context"
	"errors"
	"fmt"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/resource"
	"github.com/memochou1993/gh-rankings/app/response"
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

func (u *userWorker) Collect() error {
	logger.Info("Collecting users...")
	u.From = time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	u.To = time.Now()

	if u.Worker.Timestamp.IsZero() {
		if user := u.UserModel.FindLast(); user.ID() != "" {
			u.From = user.CreatedAt.AddDate(0, 0, -7).Truncate(24 * time.Hour)
		}
	}

	return u.Travel()
}

func (u *userWorker) Travel() error {
	if u.From.After(u.To) {
		return nil
	}

	var users []model.User
	u.SearchQuery.SearchArguments.Query = u.buildSearchQuery()
	logger.Debug(fmt.Sprintf("User Query: %s", u.SearchQuery.SearchArguments.Query))
	if err := u.Fetch(&users); err != nil {
		return err
	}

	if res := u.UserModel.Store(users); res != nil {
		if res.ModifiedCount > 0 {
			logger.Success(fmt.Sprintf("Updated %d users!", res.ModifiedCount))
		}
		if res.UpsertedCount > 0 {
			logger.Success(fmt.Sprintf("Inserted %d users!", res.UpsertedCount))
		}
	}
	for _, user := range users {
		if err := u.Update(user); err != nil {
			return err
		}
	}
	u.From = u.From.AddDate(0, 0, 7)

	return u.Travel()
}

func (u *userWorker) Fetch(users *[]model.User) error {
	res := response.User{}
	if err := u.query(*u.SearchQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		*users = append(*users, edge.Node)
	}
	res.Data.RateLimit.Break(collecting)
	if !res.Data.Search.PageInfo.HasNextPage {
		u.SearchQuery.SearchArguments.After = ""
		return nil
	}
	u.SearchQuery.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return u.Fetch(users)
}

func (u *userWorker) Update(user model.User) error {
	u.GistQuery.Field = model.TypeUser
	u.GistQuery.OwnerArguments.Login = strconv.Quote(user.ID())
	if err := u.UpdateGists(user); err != nil {
		return err
	}

	u.RepositoryQuery.Field = model.TypeUser
	u.RepositoryQuery.OwnerArguments.Login = strconv.Quote(user.ID())
	if err := u.UpdateRepositories(user); err != nil {
		return err
	}

	return nil
}

func (u *userWorker) UpdateGists(user model.User) error {
	var gists []model.Gist
	if err := u.FetchGists(&gists); err != nil {
		return err
	}
	u.UserModel.UpdateGists(user, gists)
	logger.Success(fmt.Sprintf("Updated %d %s gists!", len(gists), model.TypeUser))
	return nil
}

func (u *userWorker) UpdateRepositories(user model.User) error {
	var repositories []model.Repository
	if err := u.FetchRepositories(&repositories); err != nil {
		return err
	}
	u.UserModel.UpdateRepositories(user, repositories)
	logger.Success(fmt.Sprintf("Updated %d %s repositories!", len(repositories), model.TypeUser))
	return nil
}

func (u *userWorker) FetchGists(gists *[]model.Gist) error {
	res := response.User{}
	if err := u.query(*u.GistQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.User.Gists.Edges {
		*gists = append(*gists, edge.Node)
	}
	res.Data.RateLimit.Break(collecting)
	if !res.Data.User.Gists.PageInfo.HasNextPage {
		u.GistQuery.GistsArguments.After = ""
		return nil
	}
	u.GistQuery.GistsArguments.After = strconv.Quote(res.Data.User.Gists.PageInfo.EndCursor)

	return u.FetchGists(gists)
}

func (u *userWorker) FetchRepositories(repositories *[]model.Repository) error {
	res := response.User{}
	if err := u.query(*u.RepositoryQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.User.Repositories.Edges {
		*repositories = append(*repositories, edge.Node)
	}
	res.Data.RateLimit.Break(collecting)
	if !res.Data.User.Repositories.PageInfo.HasNextPage {
		u.RepositoryQuery.RepositoriesArguments.After = ""
		return nil
	}
	u.RepositoryQuery.RepositoriesArguments.After = strconv.Quote(res.Data.User.Repositories.PageInfo.EndCursor)

	return u.FetchRepositories(repositories)
}

func (u *userWorker) Rank() {
	logger.Info("Executing user rank pipelines...")
	var pipelines []*model.Pipeline
	pipelines = append(pipelines, u.buildRankPipelines()...)

	ch := make(chan struct{}, 2)
	wg := sync.WaitGroup{}
	wg.Add(len(pipelines))

	now := time.Now()
	for i, p := range pipelines {
		ch <- struct{}{}
		go func(p *model.Pipeline) {
			defer wg.Done()
			RankModel.Store(u.UserModel, *p, now)
			<-ch
		}(p)
		if (i+1)%10 == 0 || (i+1) == len(pipelines) {
			logger.Success(fmt.Sprintf("Executed %d of %d user rank pipelines!", i+1, len(pipelines)))
		}
	}
	wg.Wait()
	u.Worker.seal(TimestampUserRanks, now)

	tags := []string{fmt.Sprintf("type:%s", model.TypeUser), fmt.Sprintf("type:%s", model.TypeOrganization)}
	RankModel.Delete(now, tags...)
}

func (u *userWorker) query(q model.Query, res *response.User) (err error) {
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
		return u.query(q, res)
	}
	return
}

func (u *userWorker) buildSearchQuery() string {
	from := u.From.Format(time.RFC3339)
	to := u.From.AddDate(0, 0, 7).Format(time.RFC3339)
	q := model.SearchQuery{
		Created:   fmt.Sprintf("%s..%s", from, to),
		Followers: ">=100",
		Sort:      "joined-asc",
		Type:      model.TypeUser,
	}
	return strconv.Quote(util.ParseStruct(q, " "))
}

func (u *userWorker) buildRankPipelines() (pipelines []*model.Pipeline) {
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
		pipelines = append(pipelines, u.buildRankPipeline(field, tag))
		pipelines = append(pipelines, u.buildRankPipelinesByLocation(field, tag)...)
	}
	pipelines = append(pipelines, u.buildRepositoryRankPipelinesByLanguage("forks", tag)...)
	pipelines = append(pipelines, u.buildRepositoryRankPipelinesByLanguage("stargazers", tag)...)
	pipelines = append(pipelines, u.buildRepositoryRankPipelinesByLanguage("watchers", tag)...)
	return
}

func (u *userWorker) buildRankPipeline(field string, tags ...string) *model.Pipeline {
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

func (u *userWorker) buildRankPipelinesByLocation(field string, tags ...string) (pipelines []*model.Pipeline) {
	for _, location := range resource.Locations {
		pipelines = append(pipelines, u.buildRankPipeline(field, append(tags, fmt.Sprintf("location:%s", location.Name))...))
		for _, city := range location.Cities {
			pipelines = append(pipelines, u.buildRankPipeline(field, append(tags, fmt.Sprintf("location:%s, %s", city.Name, location.Name))...))
		}
	}
	return
}

func (u *userWorker) buildRepositoryRankPipelinesByLanguage(field string, tags ...string) (pipelines []*model.Pipeline) {
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
