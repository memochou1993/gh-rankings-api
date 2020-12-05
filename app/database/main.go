package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

var client *mongo.Client

func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	opts := options.Client().ApplyURI(os.Getenv("DB_HOST"))
	client, err = mongo.Connect(ctx, opts)

	if err != nil {
		log.Fatalln(err.Error())
	}
}

func CreateIndexes(ctx context.Context, collection string, keys []string) ([]string, error) {
	c := getCollection(collection)

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

func getCollection(name string) *mongo.Collection {
	return client.Database(os.Getenv("DB_DATABASE")).Collection(name)
}
