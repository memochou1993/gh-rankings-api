package model

import (
	"context"
	"fmt"
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"log"
	"strings"
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

func (u *Users) CollectUsers() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	count, err := database.Count(ctx, CollectionUsers)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if err := u.SearchUsers(ctx); err != nil {
		return err
	}
	if err := u.StoreUsers(ctx); err != nil {
		return err
	}

	return database.CreateIndexes(ctx, CollectionUsers, []string{"name"})
}

func (u *Users) SearchUsers(ctx context.Context) error {
	args := &SearchArguments{
		First: 100,
		Query: "\"repos:>=5 followers:>=10\"",
		Type:  "USER",
	}

	return app.Fetch(ctx, []byte(u.getQuery(args)), u)
}

func (u *Users) StoreUsers(ctx context.Context) error {
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

func (u *Users) getQuery(args *SearchArguments) string {
	data, err := ioutil.ReadFile(fmt.Sprintf("./app/query/%s.graphql", SearchUsers))
	if err != nil {
		log.Fatal(err.Error())
	}

	return strings.Replace(string(data), "<args>", joinArguments(args), 1)
}
