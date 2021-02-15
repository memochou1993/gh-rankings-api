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
	"github.com/memochou1993/gh-rankings/util"
	"os"
	"strconv"
	"sync"
	"time"
)

type User struct {
	*Worker
	From            time.Time
	To              time.Time
	UserModel       *model.UserModel
	RankModel       *model.RankModel
	SearchQuery     *query.Query
	GistQuery       *query.Query
	RepositoryQuery *query.Query
}

func (u *User) Init() {
	u.Worker.load(TimestampUser)
}

func (u *User) Collect() error {
	logger.Info("Collecting users...")
	u.From = time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	u.To = time.Now()

	if u.Worker.Timestamp.IsZero() {
		last := model.User{}
		if u.UserModel.Model.Last(&last); last.ID() != "" {
			u.From = last.CreatedAt.AddDate(0, 0, -7).Truncate(24 * time.Hour)
		}
	}

	return u.Travel()
}

func (u *User) Travel() error {
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

func (u *User) Fetch(users *[]model.User) error {
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

func (u *User) Update(user model.User) error {
	u.GistQuery.Type = model.TypeUser
	u.GistQuery.OwnerArguments.Login = strconv.Quote(user.ID())
	if err := u.UpdateGists(user); err != nil {
		return err
	}

	u.RepositoryQuery.Type = model.TypeUser
	u.RepositoryQuery.OwnerArguments.Login = strconv.Quote(user.ID())
	if err := u.UpdateRepositories(user); err != nil {
		return err
	}

	return nil
}

func (u *User) UpdateGists(user model.User) error {
	var gists []query.Gist
	if err := u.FetchGists(&gists); err != nil {
		return err
	}
	u.UserModel.UpdateGists(user, gists)
	logger.Success(fmt.Sprintf("Updated %d %s gists!", len(gists), model.TypeUser))
	return nil
}

func (u *User) UpdateRepositories(user model.User) error {
	var repositories []model.Repository
	if err := u.FetchRepositories(&repositories); err != nil {
		return err
	}
	u.UserModel.UpdateRepositories(user, repositories)
	logger.Success(fmt.Sprintf("Updated %d %s repositories!", len(repositories), model.TypeUser))
	return nil
}

func (u *User) FetchGists(gists *[]query.Gist) error {
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

func (u *User) FetchRepositories(repositories *[]model.Repository) error {
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

func (u *User) Rank() {
	logger.Info("Executing user rank pipelines...")
	pipelines := u.buildRankPipelines()

	ch := make(chan struct{}, 2)
	wg := sync.WaitGroup{}
	wg.Add(len(pipelines))

	now := time.Now()
	for i, p := range pipelines {
		ch <- struct{}{}
		go func(p *pipeline.Pipeline) {
			defer wg.Done()
			u.RankModel.Store(u.UserModel, *p, now)
			<-ch
		}(p)
		if (i+1)%10 == 0 || (i+1) == len(pipelines) {
			logger.Success(fmt.Sprintf("Executed %d of %d user rank pipelines!", i+1, len(pipelines)))
		}
	}
	wg.Wait()
	u.Worker.save(TimestampUser, now)

	u.RankModel.Delete(now, model.TypeUser)
}

func (u *User) query(q query.Query, res *response.User) (err error) {
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

func (u *User) buildSearchQuery() string {
	from := u.From.Format(time.RFC3339)
	to := u.From.AddDate(0, 0, 7).Format(time.RFC3339)
	q := query.SearchQuery{
		Created:   fmt.Sprintf("%s..%s", from, to),
		Followers: "100..*",
		Sort:      "joined-asc",
		Type:      model.TypeUser,
	}
	return strconv.Quote(util.ParseStruct(q, " "))
}

func (u *User) buildRankPipelines() (pipelines []*pipeline.Pipeline) {
	rankType := model.TypeUser
	fields := []string{
		"followers",
		"gists.forks",
		"gists.stargazers",
		"repositories.forks",
		"repositories.stargazers",
		"repositories.watchers",
	}
	for _, field := range fields {
		pipelines = append(pipelines, pipeline.RankByField(rankType, field))
		pipelines = append(pipelines, pipeline.RankByLocation(rankType, field)...)
	}
	pipelines = append(pipelines, pipeline.RankOwnerRepositoryByLanguage(rankType, "repositories.stargazers")...)
	pipelines = append(pipelines, pipeline.RankOwnerRepositoryByLanguage(rankType, "repositories.forks")...)
	pipelines = append(pipelines, pipeline.RankOwnerRepositoryByLanguage(rankType, "repositories.watchers")...)
	return
}

func NewUserWorker() *User {
	return &User{
		Worker:          &Worker{},
		UserModel:       model.NewUserModel(),
		RankModel:       model.NewRankModel(),
		SearchQuery:     query.Owners(),
		GistQuery:       query.OwnerGists(),
		RepositoryQuery: query.OwnerRepositories(),
	}
}
