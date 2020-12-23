package model

import (
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	TypeRepository   = "repository"
	TypeUser         = "user"
	TypeOrganization = "organization"
)

type Interface interface {
	Name() string
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
