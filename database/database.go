package database

import (
	"context"
	"github.com/memochou1993/github-rankings/logger"
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
	count, err := Collection(collection).CountDocuments(context.Background(), bson.D{})
	if err != nil {
		log.Fatalln(err.Error())
	}
	return count
}

func UpdateOne(collection string, filter bson.D, update bson.D, opts ...*options.UpdateOptions) {
	if _, err := Collection(collection).UpdateOne(context.Background(), filter, update, opts...); err != nil {
		log.Fatalln(err.Error())
	}
}

func UpdateMany(collection string, filter bson.D, update bson.D, opts ...*options.UpdateOptions) {
	if _, err := Collection(collection).UpdateMany(context.Background(), filter, update, opts...); err != nil {
		logger.Error(err)
	}
}

func All(ctx context.Context, collection string) *mongo.Cursor {
	opts := options.Find().SetBatchSize(1000)
	cursor, err := Collection(collection).Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatalln(err.Error())
	}
	return cursor
}

func Aggregate(ctx context.Context, collection string, pipeline []bson.D) *mongo.Cursor {
	opts := options.Aggregate().SetBatchSize(1000)
	cursor, err := Collection(collection).Aggregate(ctx, pipeline, opts)
	if err != nil {
		log.Fatalln(err.Error())
	}
	return cursor
}

func CloseCursor(ctx context.Context, cursor *mongo.Cursor) {
	if err := cursor.Close(ctx); err != nil {
		log.Fatalln(err.Error())
	}
}

func Indexes(collection string) (indexes []bson.D) {
	ctx := context.Background()
	cursor, err := Collection(collection).Indexes().List(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Fatalln(err.Error())
		}
	}()
	if err := cursor.All(ctx, &indexes); err != nil {
		log.Fatalln(err.Error())
	}
	return
}

func CreateIndexes(collection string, keys []string) {
	var models []mongo.IndexModel
	for _, key := range keys {
		models = append(models, mongo.IndexModel{
			Keys:    bson.D{{key, 1}},
			Options: options.Index().SetName(key),
		})
	}
	if _, err := Collection(collection).Indexes().CreateMany(context.Background(), models); err != nil {
		log.Fatalln(err.Error())
	}
}

func CreateUniqueIndexes(collection string, keys []string) {
	var models []mongo.IndexModel
	for _, key := range keys {
		models = append(models, mongo.IndexModel{
			Keys:    bson.D{{key, 1}},
			Options: options.Index().SetUnique(true).SetName(key),
		})
	}
	if _, err := Collection(collection).Indexes().CreateMany(context.Background(), models); err != nil {
		log.Fatalln(err.Error())
	}
}
