package model

import (
	"context"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/pipeline"
	"github.com/memochou1993/gh-rankings/app/query"
	"github.com/memochou1993/gh-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
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

func (r *RepositoryModel) List(req *request.Repository) (repositories []Repository) {
	ctx := context.Background()

	p := pipeline.ListRepositories(req)
	if req.Q != "" {
		p = pipeline.SearchRepositories(req)
	}

	cursor := database.Aggregate(ctx, r.Model.Name(), p)
	repositories = make([]Repository, req.Limit)
	if err := cursor.All(ctx, &repositories); err != nil {
		log.Fatal(err.Error())
	}

	return
}

func (r *RepositoryModel) FindByName(name string) (repository Repository) {
	r.Model.FindByName(name, &repository)
	return
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
		Model: &Model{
			name: "repositories",
		},
	}
}
