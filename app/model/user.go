package model

import (
	"context"
	"fmt"
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

type UserCollection struct {
	Collection
	Response struct {
		Data struct {
			Search struct {
				UserCount int `json:"userCount"`
				Edges     []struct {
					Cursor string `json:"cursor"`
					Node   User   `json:"node"`
				} `json:"edges"`
				PageInfo app.PageInfo `json:"pageInfo"`
			} `json:"search"`
			User struct {
				AvatarURL string    `json:"avatarUrl"`
				CreatedAt time.Time `json:"createdAt"`
				Email     string    `json:"email"`
				Followers struct {
					TotalCount int `json:"totalCount"`
				} `json:"followers"`
				Location     string `json:"location"`
				Login        string `json:"login"`
				Name         string `json:"name"`
				Repositories struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Cursor string `json:"cursor"`
						Node   struct {
							Name            string `json:"name"`
							PrimaryLanguage struct {
								Name string `json:"name"`
							} `json:"primaryLanguage"`
							Stargazers struct {
								TotalCount int `json:"totalCount"`
							} `json:"stargazers"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo app.PageInfo `json:"pageInfo"`
				} `json:"repositories"`
			} `json:"user"`
			RateLimit app.RateLimit `json:"rateLimit"`
		} `json:"data"`
		Errors []app.Error `json:"errors"`
	}
}

type User struct {
	AvatarURL string    `json:"avatarUrl"`
	CreatedAt time.Time `json:"createdAt"`
	Email     string    `json:"email"`
	Followers struct {
		TotalCount int `json:"totalCount"`
	} `json:"followers"`
	Location     string `json:"location"`
	Login        string `json:"login"`
	Name         string `json:"name"`
	Repositories []struct {
		Name            string `json:"name"`
		PrimaryLanguage struct {
			Name string `json:"name"`
		} `json:"primaryLanguage"`
		Stargazers struct {
			TotalCount int `json:"totalCount"`
		} `json:"stargazers"`
	} `json:"repositories"`
}

func (u *UserCollection) Init() error {
	u.SetCollectionName("users")

	if u.Count() > 0 {
		return nil
	}
	if err := u.Index([]string{"login"}); err != nil {
		return err
	}

	return nil
}

func (u *UserCollection) Collect() error {
	if u.Count() > 0 {
		return nil
	}

	date := time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	r := app.Request{
		Schema: app.Read("users"),
		SearchArguments: app.SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
	if err := u.Travel(&date, &r); err != nil {
		return nil
	}

	return nil
}

func (u *UserCollection) Travel(t *time.Time, r *app.Request) error {
	layout := "2006-01-02"
	if t.After(time.Now()) {
		return nil
	}
	q := app.ArgumentsQuery{
		Created:   r.Range(t.Format(layout), t.AddDate(0, 0, 6).Format(layout)),
		Followers: ">=10",
		Repos:     ">=5",
	}
	r.SearchArguments.Query = r.String(util.JoinStruct(q, " "))

	var users []interface{}
	if err := u.FetchUsers(r, &users); err != nil {
		return err
	}
	if err := u.StoreUsers(users); err != nil {
		return err
	}
	*t = t.AddDate(0, 0, 7)

	return u.Travel(t, r)
}

func (u *UserCollection) FetchUsers(r *app.Request, users *[]interface{}) error {
	logger.Debug(r.SearchArguments)
	logger.Info("Searching users...")
	if err := u.Fetch(r); err != nil {
		return err
	}
	count := len(u.Response.Data.Search.Edges)
	if count == 0 {
		return nil
	}
	for _, edge := range u.Response.Data.Search.Edges {
		*users = append(*users, edge.Node)
	}
	logger.Success(fmt.Sprintf("Discovered %d users", count))

	if !u.Response.Data.Search.PageInfo.HasNextPage {
		r.SearchArguments.After = ""
		return nil
	}
	r.SearchArguments.After = r.String(u.Response.Data.Search.PageInfo.EndCursor)

	return u.FetchUsers(r, users)
}

func (u *UserCollection) StoreUsers(users []interface{}) error {
	if len(users) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := u.GetCollection().InsertMany(ctx, users)
	logger.Success(fmt.Sprintf("Stored %d users", len(users)))

	return err
}

func (u *UserCollection) Update() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cursor, err := u.GetCollection().Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Fatalln(err.Error())
		}
	}()

	r := app.Request{
		Schema: app.Read("user_repositories"),
		RepositoriesArguments: app.RepositoriesArguments{
			First:             100,
			OrderBy:           "{field:STARGAZERS,direction:DESC}",
			OwnerAffiliations: "OWNER",
		},
	}
	for cursor.Next(ctx) {
		user := User{}
		if err := cursor.Decode(&user); err != nil {
			log.Fatalln(err.Error())
		}
		r.UserArguments.Login = r.String(user.Login)

		var repositories []interface{}
		if err := u.FetchRepositories(&r, &repositories); err != nil {
			return err
		}
		u.StoreRepositories(user, repositories)
	}

	return nil
}

func (u *UserCollection) FetchRepositories(r *app.Request, repos *[]interface{}) error {
	logger.Debug(r.UserArguments)
	logger.Debug(r.RepositoriesArguments)
	logger.Info("Searching repositories...")
	if err := u.Fetch(r); err != nil {
		return err
	}
	count := len(u.Response.Data.User.Repositories.Edges)
	if count == 0 {
		return nil
	}
	for _, edge := range u.Response.Data.User.Repositories.Edges {
		*repos = append(*repos, edge.Node)
	}
	logger.Success(fmt.Sprintf("Discovered %d repositories", count))

	if !u.Response.Data.User.Repositories.PageInfo.HasNextPage {
		r.RepositoriesArguments.After = ""
		return nil
	}
	r.RepositoriesArguments.After = r.String(u.Response.Data.User.Repositories.PageInfo.EndCursor)

	return u.FetchRepositories(r, repos)
}

func (u *UserCollection) StoreRepositories(user User, repos []interface{}) {
	if len(repos) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"login", user.Login}}
	update := bson.D{{"$set", bson.D{{"repositories", repos}}}}

	u.GetCollection().FindOneAndUpdate(ctx, filter, update)
	logger.Success(fmt.Sprintf("Stored %d repositories", len(repos)))
}

func (u *UserCollection) Fetch(r *app.Request) error {
	u.Response.Data.RateLimit.Check()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Query(ctx, r, &u.Response)
	logger.Debug(u.Response.Data.RateLimit)
	for _, err := range u.Response.Errors {
		logger.Error(err.Message)
	}

	return err
}
