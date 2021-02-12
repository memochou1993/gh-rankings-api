package model

import (
	"github.com/memochou1993/gh-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

const (
	TypeUser         = "user"
	TypeOrganization = "organization"
	TypeRepository   = "repository"
)

type Interface interface {
	Name() string
	Collection() *mongo.Collection
}

type Model struct {
	name string
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) Collection() *mongo.Collection {
	return database.Collection(m.name)
}

func (m *Model) Last(v interface{}) {
	opts := options.FindOne().SetSort(bson.D{{"$natural", -1}})
	res := database.FindOne(m.Name(), bson.D{}, opts)
	if err := res.Decode(&v); err != nil && err != mongo.ErrNoDocuments {
		log.Fatal(err.Error())
	}
	return
}

// TODO: should create a NewModel method
