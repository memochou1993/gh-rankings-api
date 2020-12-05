package model

import (
	"context"
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

func (u *Users) Collect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	count, err := database.Count(ctx, CollectionUsers)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if err := u.Search(ctx); err != nil {
		return err
	}
	if err := u.Store(ctx); err != nil {
		return err
	}

	return database.CreateIndexes(ctx, CollectionUsers, []string{"name"})
}

func (u *Users) Search(ctx context.Context) error {
	args := &query.SearchArguments{
		First: 100,
		Query: "\"repos:>=5 followers:>=10\"",
		Type:  "USER",
	}

	return app.Fetch(ctx, []byte(args.Read(SearchUsers)), u)
}

func (u *Users) Store(ctx context.Context) error {
	var items []interface{}
	for _, user := range u.Data.Search.Edges {
		items = append(items, bson.M{
			"id":   user.Node.ID,
			"name": user.Node.Login,
		})
	}

	_, err := database.GetCollection(CollectionUsers).InsertMany(ctx, items)

	return err
}
