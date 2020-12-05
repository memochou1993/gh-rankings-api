package database

import (
	"context"
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/app/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func CollectInitialUsers() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	count, err := Count(ctx, "users")
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	users, err := app.SearchInitialUsers(ctx)
	if err != nil {
		return err
	}
	if _, err := StoreInitialUsers(ctx, users); err != nil {
		return err
	}

	err = CreateIndexes(ctx, "users", []string{"name"})

	return err
}

func StoreInitialUsers(ctx context.Context, users model.InitialUsers) (*mongo.InsertManyResult, error) {
	var items []interface{}
	for _, user := range users.Data.Search.Edges {
		items = append(items, bson.M{
			"id":   user.Node.ID,
			"name": user.Node.Login,
		})
	}

	return GetCollection("users").InsertMany(ctx, items)
}
