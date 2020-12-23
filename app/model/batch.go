package model

import (
	"context"
	"github.com/memochou1993/github-rankings/database"
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
	database.UpdateOne(b.Name(), filter, update, opts)
}
