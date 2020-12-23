package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
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

func (m *Model) Get(model Interface) (batch Batch) {
	filter := bson.D{{"model", model.Name()}}
	if err := m.Collection().FindOne(context.Background(), filter).Decode(&batch); err != nil {
		if err == mongo.ErrNoDocuments {
			return
		}
		log.Fatalln(err.Error())
	}

	return
}

func (b *BatchModel) Update(model Interface) {
	filter := bson.D{{"model", model.Name()}}
	update := bson.D{{"$inc", bson.D{{"batch", 1}}}}
	opts := options.Update().SetUpsert(true)
	if _, err := b.Collection().UpdateOne(context.Background(), filter, update, opts); err != nil {
		log.Fatalln(err.Error())
	}
}
