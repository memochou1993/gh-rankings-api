package app

import (
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/mongo"
)

type CollectionInterface interface {
	GetCollection() *mongo.Collection
}

type Collection struct {
	collectionName string
}

func (c *Collection) SetCollectionName(collectionName string) {
	c.collectionName = collectionName
}

func (c *Collection) GetCollection() *mongo.Collection {
	return database.GetCollection(c.collectionName)
}
