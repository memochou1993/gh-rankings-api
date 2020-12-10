package database

import (
	"context"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var client *mongo.Client

func Init() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	opts := options.Client().ApplyURI(viper.GetString("DB_HOST"))
	client, err = mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func GetDatabase() *mongo.Database {
	return client.Database(viper.GetString("DB_DATABASE"))
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
			Keys:    bson.D{{key, 1}},
			Options: options.Index().SetName(key),
		})
	}

	_, err := GetCollection(collection).Indexes().CreateMany(ctx, models)

	return err
}
