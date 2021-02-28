package worker

import (
	"context"
	"errors"
	"fmt"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/pipeline"
	"github.com/memochou1993/gh-rankings/app/query"
	"github.com/memochou1993/gh-rankings/app/resource"
	"github.com/memochou1993/gh-rankings/app/response"
	"github.com/memochou1993/gh-rankings/logger"
	"strconv"
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
	u.Worker.load(timestampUser)
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

	if err := u.Prepare(); err != nil {
		return err
	}

	return u.Travel()
}

func (u *User) Prepare() error {
	for _, user := range resource.SpecifiedUsers {
		var users []model.User
		u.SearchQuery.SearchArguments.SetQuery(query.SearchSpecifiedUser(user.Login))
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
	}

	return nil
}

func (u *User) Travel() error {
	if u.From.After(u.To) {
		return nil
	}

	var users []model.User
	u.SearchQuery.SearchArguments.SetQuery(query.SearchUsers(u.From, u.From.AddDate(0, 0, 7)))
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
	res.Data.RateLimit.Throttle(collecting)
	if !res.Data.Search.PageInfo.HasNextPage {
		u.SearchQuery.SearchArguments.After = ""
		return nil
	}
	u.SearchQuery.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return u.Fetch(users)
}

func (u *User) Update(user model.User) error {
	u.GistQuery.Type = app.TypeUser
	u.GistQuery.OwnerArguments.Login = strconv.Quote(user.ID())
	if err := u.UpdateGists(user); err != nil {
		return err
	}

	u.RepositoryQuery.Type = app.TypeUser
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
	logger.Success(fmt.Sprintf("Updated %d %s gists!", len(gists), app.TypeUser))
	return nil
}

func (u *User) UpdateRepositories(user model.User) error {
	var repositories []model.Repository
	if err := u.FetchRepositories(&repositories); err != nil {
		return err
	}
	u.UserModel.UpdateRepositories(user, repositories)
	logger.Success(fmt.Sprintf("Updated %d %s repositories!", len(repositories), app.TypeUser))
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
	res.Data.RateLimit.Throttle(collecting)
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
	res.Data.RateLimit.Throttle(collecting)
	if !res.Data.User.Repositories.PageInfo.HasNextPage {
		u.RepositoryQuery.RepositoriesArguments.After = ""
		return nil
	}
	u.RepositoryQuery.RepositoriesArguments.After = strconv.Quote(res.Data.User.Repositories.PageInfo.EndCursor)

	return u.FetchRepositories(repositories)
}

func (u *User) Rank() {
	logger.Info("Executing user rank pipelines...")
	pipelines := pipeline.RankUser()
	timestamp := time.Now()
	for i, p := range pipelines {
		u.RankModel.Store(u.UserModel, *p, timestamp)
		if (i+1)%10 == 0 || (i+1) == len(pipelines) {
			logger.Success(fmt.Sprintf("Executed %d of %d user rank pipelines!", i+1, len(pipelines)))
		}
	}
	u.Worker.save(timestampUser, timestamp)
	u.RankModel.Delete(timestamp, app.TypeUser)
}

func (u *User) query(q query.Query, res *response.User) (err error) {
	err = app.Fetch(context.Background(), fmt.Sprint(q), res)
	if res.Message != "" {
		err = errors.New(res.Message)
		res.Message = ""
	}
	for _, err := range res.Errors {
		return err
	}
	if err != nil {
		logger.Error(err.Error())
		logger.Warning("Retrying...")
		time.Sleep(10 * time.Second)
		return u.query(q, res)
	}
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
