package model

import (
	"context"
	"fmt"
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/app/query"
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
				PageInfo query.PageInfo `json:"pageInfo"`
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
					PageInfo query.PageInfo `json:"pageInfo"`
				} `json:"repositories"`
			} `json:"user"`
			RateLimit query.RateLimit `json:"rateLimit"`
		} `json:"data"`
		Errors []query.Error `json:"errors"`
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

	date := time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC)
	request := query.Request{
		Schema: query.Read("users"),
		SearchArguments: query.SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
	if err := u.Travel(&date, &request); err != nil {
		return nil
	}

	return nil
}

func (u *UserCollection) Travel(t *time.Time, r *query.Request) error {
	layout := "2006-01-02"
	if t.After(time.Now()) {
		return nil
	}
	q := query.ArgumentsQuery{
		Created:   fmt.Sprintf("%s..%s", t.Format(layout), t.AddDate(0, 0, 6).Format(layout)),
		Followers: ">=10",
		Repos:     ">=5",
	}
	r.SearchArguments.Query = q.Join()
	if err := u.FetchUsers(r); err != nil {
		return err
	}
	*t = t.AddDate(0, 0, 7)

	return u.Travel(t, r)
}

func (u *UserCollection) FetchUsers(r *query.Request) error {
	after := &r.SearchArguments.After
	endCursor := &u.Response.Data.Search.PageInfo.EndCursor
	u.Response.Data.RateLimit.Check()
	if *endCursor != "" {
		*after = fmt.Sprintf("\"%s\"", *endCursor)
	}
	util.Log("DEBUG", r.SearchArguments)
	util.Log("INFO", "Searching users...")
	if err := u.Fetch(r.Join()); err != nil {
		return err
	}
	u.LogErrors()
	util.Log("DEBUG", u.Response.Data.RateLimit)
	count := len(u.Response.Data.Search.Edges)
	if count == 0 {
		return nil
	}
	if err := u.StoreUsers(); err != nil {
		return err
	}
	util.Log("INFO", fmt.Sprintf("Discovered %d users", count))
	if !u.Response.Data.Search.PageInfo.HasNextPage {
		*endCursor = ""
		*after = ""
		return nil
	}

	return u.FetchUsers(r)
}

func (u *UserCollection) Update() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	request := query.Request{
		Schema: query.Read("user_repositories"),
		RepositoriesArguments: query.RepositoriesArguments{
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
		request.UserArguments.Login = fmt.Sprintf("\"%s\"", user.Login)
		after := &request.SearchArguments.After
		endCursor := &u.Response.Data.Search.PageInfo.EndCursor
		var repositories []interface{}
		for {
			u.Response.Data.RateLimit.Check()
			if *endCursor != "" {
				*after = fmt.Sprintf("\"%s\"", *endCursor)
			}
			util.Log("DEBUG", request.UserArguments)
			util.Log("DEBUG", request.RepositoriesArguments)
			util.Log("INFO", "Searching user repositories...")
			if err := u.Fetch(request.Join()); err != nil {
				return err
			}
			u.LogErrors()
			util.Log("DEBUG", u.Response.Data.RateLimit)
			count := len(u.Response.Data.User.Repositories.Edges)
			if count == 0 {
				break
			}
			for _, edge := range u.Response.Data.User.Repositories.Edges {
				repositories = append(repositories, edge.Node)
			}
			util.Log("INFO", fmt.Sprintf("Discovered %d repositories", count))
			if !u.Response.Data.User.Repositories.PageInfo.HasNextPage {
				*endCursor = ""
				*after = ""
				break
			}
		}
		u.UpdateRepositories(user, repositories)
	}

	return nil
}

func (u *UserCollection) StoreUsers() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var documents []interface{}
	for _, edge := range u.Response.Data.Search.Edges {
		documents = append(documents, edge.Node)
	}

	_, err := u.GetCollection().InsertMany(ctx, documents)

	return err
}

func (u *UserCollection) UpdateRepositories(user User, documents []interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"login", user.Login}}
	update := bson.D{{"$set", bson.D{{"repositories", documents}}}}

	u.GetCollection().FindOneAndUpdate(ctx, filter, update)
}

func (u *UserCollection) Fetch(q []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return app.Fetch(ctx, q, &u.Response)
}

func (u *UserCollection) LogErrors() {
	if len(u.Response.Errors) > 0 {
		for _, err := range u.Response.Errors {
			util.Log("ERROR", err.Message)
		}
	}
}
