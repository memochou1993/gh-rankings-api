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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	opts := options.Client().ApplyURI(viper.GetString("DB_HOST"))
	client, err = mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatal(err.Error())
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
		log.Fatal(err.Error())
	}
	return count
}

func BulkWrite(collection string, models []mongo.WriteModel) *mongo.BulkWriteResult {
	res, err := Collection(collection).BulkWrite(context.Background(), models)
	if err != nil {
		log.Fatal(err.Error())
	}
	return res
}

func FindOne(collection string, filter bson.D, opts ...*options.FindOneOptions) *mongo.SingleResult {
	return Collection(collection).FindOne(context.Background(), filter, opts...)
}

func UpdateOne(collection string, filter bson.D, update bson.D, opts ...*options.UpdateOptions) {
	if _, err := Collection(collection).UpdateOne(context.Background(), filter, update, opts...); err != nil {
		log.Fatal(err.Error())
	}
}

func UpdateMany(collection string, filter bson.D, update bson.D, opts ...*options.UpdateOptions) {
	if _, err := Collection(collection).UpdateMany(context.Background(), filter, update, opts...); err != nil {
		log.Fatal(err.Error())
	}
}

func DeleteMany(collection string, filter bson.D, opts ...*options.DeleteOptions) {
	if _, err := Collection(collection).DeleteMany(context.Background(), filter, opts...); err != nil {
		log.Fatal(err.Error())
	}
}

func All(ctx context.Context, collection string, skip int, limit int) *mongo.Cursor {
	opts := options.Find().SetBatchSize(1000).SetSkip(int64(skip)).SetLimit(int64(limit))
	cursor, err := Collection(collection).Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatal(err.Error())
	}
	return cursor
}

func Aggregate(ctx context.Context, collection string, pipeline []bson.D) *mongo.Cursor {
	opts := options.Aggregate().SetBatchSize(1000).SetAllowDiskUse(true)
	cursor, err := Collection(collection).Aggregate(ctx, pipeline, opts)
	if err != nil {
		log.Fatal(err.Error())
	}
	return cursor
}

func CloseCursor(ctx context.Context, cursor *mongo.Cursor) {
	if err := cursor.Close(ctx); err != nil {
		log.Fatal(err.Error())
	}
}

func Indexes(collection string) (indexes []bson.D) {
	ctx := context.Background()
	cursor, err := Collection(collection).Indexes().List(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer CloseCursor(ctx, cursor)
	if err := cursor.All(ctx, &indexes); err != nil {
		log.Fatal(err.Error())
	}
	return
}

func CreateIndexes(collection string, keys []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var models []mongo.IndexModel
	for _, key := range keys {
		models = append(models, mongo.IndexModel{
			Keys:    bson.D{{key, 1}},
			Options: options.Index().SetName(key),
		})
	}
	if _, err := Collection(collection).Indexes().CreateMany(ctx, models); err != nil {
		log.Fatal(err.Error())
	}
}

func CreateUniqueIndexes(collection string, keys []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var models []mongo.IndexModel
	for _, key := range keys {
		models = append(models, mongo.IndexModel{
			Keys:    bson.D{{key, 1}},
			Options: options.Index().SetUnique(true).SetName(key),
		})
	}
	if _, err := Collection(collection).Indexes().CreateMany(ctx, models); err != nil {
		log.Fatal(err.Error())
	}
}
