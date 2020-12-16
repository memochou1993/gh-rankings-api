package app

import (
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/mongo"
)

type ModelInterface interface {
	Collection() *mongo.Collection
}

type Model struct {
	name string
}

func (m *Model) SetName(name string) {
	m.name = name
}

func (m Model) Name() string {
	return m.name
}

func (m Model) Collection() *mongo.Collection {
	return database.Collection(m.name)
}
