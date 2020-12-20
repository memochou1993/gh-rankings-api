package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type Batch struct {
	Model string `bson:"model"`
	Batch int    `bson:"batch"`
}

type BatchModel struct {
	*Model
}

func NewBatchModel() *BatchModel {
	return &BatchModel{
		&Model{
			name: "batches",
		},
	}
}

func (m *Model) Get(model string) (batch Batch) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"model", model}}
	if err := m.Collection().FindOne(ctx, filter).Decode(&batch); err != nil {
		if err == mongo.ErrNoDocuments {
			return
		}
		log.Fatalln(err.Error())
	}

	return
}

func (b *BatchModel) Update(model string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"model", model}}
	update := bson.D{{"$inc", bson.D{{"batch", 1}}}}
	opts := options.Update().SetUpsert(true)
	if _, err := b.Collection().UpdateOne(ctx, filter, update, opts); err != nil {
		log.Fatalln(err.Error())
	}
}
