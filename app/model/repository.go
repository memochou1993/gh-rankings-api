package model

import (
	"github.com/memochou1993/gh-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Repository struct {
	CreatedAt     *time.Time `json:"createdAt,omitempty" bson:"created_at"`
	Forks         *Items     `json:"forks,omitempty" bson:"forks"`
	Name          string     `json:"name,omitempty" bson:"name"`
	NameWithOwner string     `json:"nameWithOwner" bson:"_id"`
	ImageUrl      string     `json:"imageUrl,omitempty" bson:"image_url"`
	Owner         struct {
		Login string `json:"login,omitempty" bson:"login"`
	} `json:"owner,omitempty" bson:"owner"`
	PrimaryLanguage struct {
		Name string `json:"name,omitempty" bson:"name"`
	} `json:"primaryLanguage,omitempty" bson:"primary_language"`
	Stargazers *Items `json:"stargazers,omitempty" bson:"stargazers"`
	Watchers   *Items `json:"watchers,omitempty" bson:"watchers"`
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
		&Model{
			name: "repositories",
		},
	}
}
