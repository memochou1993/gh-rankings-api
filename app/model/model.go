package model

import (
	"context"
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type Interface interface {
	Collection() *mongo.Collection
}

type Model struct {
	name string
}

func (m Model) Name() string {
	return m.name
}

func (m Model) Collection() *mongo.Collection {
	return database.Collection(m.name)
}

func (m *Model) All(ctx context.Context) *mongo.Cursor {
	opts := options.Find().SetBatchSize(1000)
	cursor, err := m.Collection().Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return cursor
}
