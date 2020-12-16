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

func Database() *mongo.Database {
	return client.Database(viper.GetString("DB_DATABASE"))
}

func Collection(name string) *mongo.Collection {
	return Database().Collection(name)
}

func Count(collection string) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := Collection(collection).CountDocuments(ctx, bson.D{})
	if err != nil {
		log.Fatalln(err.Error())
	}

	return count
}

func Indexes(collection string) []bson.D {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := Collection(collection).Indexes().List(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Fatalln(err.Error())
		}
	}()

	var indexes []bson.D
	if err := cursor.All(ctx, &indexes); err != nil {
		log.Fatalln(err.Error())
	}

	return indexes
}

func CreateIndexes(collection string, keys []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var models []mongo.IndexModel
	for _, key := range keys {
		models = append(models, mongo.IndexModel{
			Keys:    bson.D{{key, 1}},
			Options: options.Index().SetName(key),
		})
	}

	_, err := Collection(collection).Indexes().CreateMany(ctx, models)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func CreateUniqueIndexes(collection string, keys []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var models []mongo.IndexModel
	for _, key := range keys {
		models = append(models, mongo.IndexModel{
			Keys:    bson.D{{key, 1}},
			Options: options.Index().SetUnique(true).SetName(key),
		})
	}

	_, err := Collection(collection).Indexes().CreateMany(ctx, models)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
