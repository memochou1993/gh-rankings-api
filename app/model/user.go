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
			User      User            `json:"user"`
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

	layout := "2006-01-02"
	date := time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	for ; date.Before(time.Now()); date.AddDate(0, 0, 7) {
		created := fmt.Sprintf("%s..%s", date.Format(layout), date.AddDate(0, 0, 6).Format(layout))
		followers := ">=10"
		repos := ">=5"
		args := query.Arguments{
			SearchArguments: query.SearchArguments{
				First: 100,
				Query: fmt.Sprintf("\"created:%s followers:%s repos:%s\"", created, followers, repos),
				Type:  "USER",
			},
		}
		for {
			u.Response.Data.RateLimit.Check()
			if u.Response.Data.Search.PageInfo.EndCursor != "" {
				args.SearchArguments.After = fmt.Sprintf("\"%s\"", u.Response.Data.Search.PageInfo.EndCursor)
			}
			util.LogStruct("Search Arguments", args.SearchArguments)
			if err := u.FetchUsers(&args); err != nil {
				return err
			}
			if len(u.Response.Errors) > 0 {
				util.LogStruct("Errors", u.Response.Errors)
			}
			util.LogStruct("Rate Limit", u.Response.Data.RateLimit)
			if len(u.Response.Data.Search.Edges) == 0 {
				break
			}
			if err := u.StoreUsers(); err != nil {
				return err
			}
			log.Println(fmt.Sprintf("Discovered %d users", len(u.Response.Data.Search.Edges)))
			if !u.Response.Data.Search.PageInfo.HasNextPage {
				u.Response.Data.Search.PageInfo.EndCursor = ""
				break
			}
		}
		date = date.AddDate(0, 0, 7)
	}

	return nil
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

	for cursor.Next(ctx) {
		user := User{}
		if err := cursor.Decode(&user); err != nil {
			log.Fatalln(err.Error())
		}

		args := query.Arguments{
			UserArguments: query.UserArguments{
				Login: fmt.Sprintf("\"%s\"", user.Login),
			},
			RepositoriesArguments: query.RepositoriesArguments{
				First:             100,
				OrderBy:           "{field: STARGAZERS, direction: DESC}",
				OwnerAffiliations: "OWNER",
			},
		}
		if err := u.FetchRepositories(&args); err != nil {
			return err
		}

		// TODO
		log.Print(u.Response.Data.User)
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

func (u *UserCollection) FetchUsers(args *query.Arguments) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return app.Fetch(ctx, args.Read("users"), &u.Response)
}

func (u *UserCollection) FetchRepositories(args *query.Arguments) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return app.Fetch(ctx, args.Read("user_repositories"), &u.Response)
}
