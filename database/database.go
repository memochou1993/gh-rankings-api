package database

import (
	"context"
	"github.com/memochou1993/github-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

var client *mongo.Client

func init() {
	util.LoadEnv()
	initClient()
}

func initClient() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	opts := options.Client().ApplyURI(os.Getenv("DB_HOST"))
	client, err = mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func GetDatabase() *mongo.Database {
	return client.Database(os.Getenv("DB_DATABASE"))
}

func GetCollection(name string) *mongo.Collection {
	return GetDatabase().Collection(name)
}

func Count(ctx context.Context, name string) (int64, error) {
	return GetCollection(name).CountDocuments(ctx, bson.M{})
}

func CreateIndexes(ctx context.Context, collection string, keys []string) error {
	var models []mongo.IndexModel
	for _, key := range keys {
		models = append(models, mongo.IndexModel{
			Keys:    bson.M{key: 1},
			Options: options.Index().SetName(key),
		})
	}

	_, err := GetCollection(collection).Indexes().CreateMany(ctx, models)

	return err
}
