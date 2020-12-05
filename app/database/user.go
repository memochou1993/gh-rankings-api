package database

import (
	"context"
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/app/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func CollectUsers() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	count, err := Count(ctx, model.CollectionUsers)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	users, err := app.SearchUsers(ctx)
	if err != nil {
		return err
	}
	if _, err := StoreUsers(ctx, users); err != nil {
		return err
	}

	err = CreateIndexes(ctx, model.CollectionUsers, []string{"name"})

	return err
}

func StoreUsers(ctx context.Context, users model.Users) (*mongo.InsertManyResult, error) {
	var items []interface{}
	for _, user := range users.Data.Search.Edges {
		items = append(items, bson.M{
			"id":   user.Node.ID,
			"name": user.Node.Login,
		})
	}

	return GetCollection(model.CollectionUsers).InsertMany(ctx, items)
}
