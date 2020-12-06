package model

import (
	"context"
	"fmt"
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/app/database"
	"github.com/memochou1993/github-rankings/app/query"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

const (
	CollectionUsers = "users"
	SearchUsers     = "search_users"
)

type Users struct {
	Data struct {
		Search struct {
			UserCount int `json:"userCount" bson:"userCount"`
			Edges     []struct {
				Cursor string `json:"cursor"`
				Node   struct {
					ID    string `json:"id"`
					Login string `json:"login"`
				} `json:"node"`
			} `json:"edges"`
			PageInfo struct {
				EndCursor   string `json:"endCursor"`
				HasNextPage bool   `json:"hasNextPage"`
				StartCursor string `json:"startCursor"`
			} `json:"pageInfo"`
		} `json:"search"`
	} `json:"data"`
}

func (u *Users) Init() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	count, err := database.Count(ctx, CollectionUsers)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if err := u.Collect(ctx); err != nil {
		return err
	}
	if err := u.Index(ctx); err != nil {
		return err
	}

	return nil
}

func (u *Users) Collect(ctx context.Context) error {
	args := &query.SearchArguments{
		First: 100,
		Query: "\"repos:>=5 followers:>=10\"",
		Type:  "USER",
	}
	u.Data.Search.PageInfo.HasNextPage = true
	for u.Data.Search.PageInfo.HasNextPage {
		if u.Data.Search.PageInfo.EndCursor != "" {
			args.After = fmt.Sprintf("\"%s\"", u.Data.Search.PageInfo.EndCursor)
		}
		if err := u.Search(ctx, args); err != nil {
			return err
		}
		if err := u.Store(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (u *Users) Search(ctx context.Context, args *query.SearchArguments) error {
	return app.Fetch(ctx, []byte(args.Read(SearchUsers)), u)
}

func (u *Users) Store(ctx context.Context) error {
	var items []interface{}
	for _, user := range u.Data.Search.Edges {
		items = append(items, bson.M{
			"_id":  user.Node.ID,
			"name": user.Node.Login,
		})
	}

	_, err := database.GetCollection(CollectionUsers).InsertMany(ctx, items)

	return err
}

func (u *Users) Index(ctx context.Context) error {
	return database.CreateIndexes(ctx, CollectionUsers, []string{"name"})
}
