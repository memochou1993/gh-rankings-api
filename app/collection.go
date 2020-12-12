package app

import (
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/mongo"
)

type CollectionInterface interface {
	GetCollection() *mongo.Collection
}

type Collection struct {
	name string
}

func (c *Collection) SetName(name string) {
	c.name = name
}

func (c *Collection) GetName() string {
	return c.name
}

func (c *Collection) GetCollection() *mongo.Collection {
	return database.GetCollection(c.name)
}
