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

func Count(collection string) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := GetCollection(collection).CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatalln(err.Error())
	}

	return count
}

func Get(collection string, opts *options.FindOneOptions) *mongo.SingleResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return GetCollection(collection).FindOne(ctx, bson.D{}, opts)
}

func GetIndexes(collection string) []bson.M {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := GetCollection(collection).Indexes().List(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Fatalln(err.Error())
		}
	}()

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		log.Fatalln(err.Error())
	}

	return indexes
}

func CreateIndexes(collection string, keys []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

func CreateUniqueIndexes(collection string, keys []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var models []mongo.IndexModel
	for _, key := range keys {
		models = append(models, mongo.IndexModel{
			Keys:    bson.D{{key, 1}},
			Options: options.Index().SetUnique(true).SetName(key),
		})
	}

	_, err := GetCollection(collection).Indexes().CreateMany(ctx, models)

	return err
}
