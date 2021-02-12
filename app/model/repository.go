package model

import (
	"github.com/memochou1993/gh-rankings/app/query"
	"github.com/memochou1993/gh-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Repository struct {
	CreatedAt     *time.Time   `json:"createdAt" bson:"created_at"`
	Forks         *query.Items `json:"forks" bson:"forks"`
	Name          string       `json:"name" bson:"name"`
	NameWithOwner string       `json:"nameWithOwner" bson:"_id"`
	ImageUrl      string       `json:"imageUrl" bson:"image_url"`
	Owner         struct {
		Login string `json:"login" bson:"login"`
	} `json:"owner" bson:"owner"`
	PrimaryLanguage struct {
		Name string `json:"name" bson:"name"`
	} `json:"primaryLanguage" bson:"primary_language"`
	Stargazers *query.Items `json:"stargazers" bson:"stargazers"`
	Watchers   *query.Items `json:"watchers" bson:"watchers"`
}

func (r *Repository) ID() string {
	return r.NameWithOwner
}

type RepositoryModel struct {
	*Model
}

func (r *RepositoryModel) Store(repositories []Repository) *mongo.BulkWriteResult {
	if len(repositories) == 0 {
		return nil
	}
	var models []mongo.WriteModel
	for _, repository := range repositories {
		filter := bson.D{{"_id", repository.ID()}}
		update := bson.D{{"$set", repository}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}
	return database.BulkWrite(r.Name(), models)
}

func NewRepositoryModel() *RepositoryModel {
	return &RepositoryModel{
		Model: NewModel("repositories"),
	}
}
