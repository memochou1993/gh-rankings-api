package app

import (
	"context"
	"github.com/memochou1993/github-rankings/app/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

type Database struct {
	Client *mongo.Client
}

func (d *Database) getClient() *mongo.Client {
	if d.Client != nil {
		return d.Client
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Client().ApplyURI(os.Getenv("DB_HOST"))
	client, err := mongo.Connect(ctx, opts)

	if err != nil {
		log.Fatalln(err.Error())
	}

	d.Client = client

	return d.Client
}

func (d *Database) StoreSearchUsers(users model.SearchUsers) (*mongo.InsertManyResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	var items []interface{}

	for _, user := range users.Data.Search.Edges {
		items = append(items, bson.M{
			"id":   user.Node.ID,
			"name": user.Node.Login,
		})
	}

	return d.getCollection("users").InsertMany(ctx, items)
}

func (d *Database) CreateIndexes(collection string, keys []string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c := d.getCollection(collection)

	var models []mongo.IndexModel

	for _, key := range keys {
		models = append(models, mongo.IndexModel{
			Keys:    bson.M{key: 1},
			Options: options.Index().SetName(key),
		})
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	return c.Indexes().CreateMany(ctx, models, opts)
}

func (d *Database) getCollection(name string) *mongo.Collection {
	return d.getClient().Database(os.Getenv("DB_DATABASE")).Collection(name)
}
