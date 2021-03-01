package model

import (
	"context"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/pipeline"
	"github.com/memochou1993/gh-rankings/app/query"
	"github.com/memochou1993/gh-rankings/app/resource"
	"github.com/memochou1993/gh-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type User struct {
	ImageUrl       string       `json:"imageUrl" bson:"image_url"`
	CreatedAt      *time.Time   `json:"createdAt" bson:"created_at"`
	Followers      *query.Items `json:"followers" bson:"followers"`
	Location       string       `json:"location" bson:"location"`
	Login          string       `json:"login" bson:"_id"`
	Name           string       `json:"name" bson:"name"`
	Gists          []query.Gist `json:"gists,omitempty" bson:"gists,omitempty"`
	Repositories   []Repository `json:"repositories,omitempty" bson:"repositories,omitempty"`
	ParsedLocation string       `json:"parsedLocation" bson:"parsed_location"`
	ParsedCity     string       `json:"parsedCity" bson:"parsed_city"`
}

func (u *User) ID() string {
	return u.Login
}

func (u *User) parseLocation() {
	u.ParsedLocation, u.ParsedCity = resource.Locate(u.Location)
}

type UserModel struct {
	*Model
}

func (u *UserModel) List(req *request.User) (users []User) {
	ctx := context.Background()

	p := pipeline.ListUsers(req)
	if req.Q != "" {
		p = pipeline.SearchUsers(req)
	}

	cursor := database.Aggregate(ctx, u.Model.Name(), p)
	users = make([]User, req.Limit)
	if err := cursor.All(ctx, &users); err != nil {
		log.Fatal(err.Error())
	}

	return
}

func (u *UserModel) FindByID(id string) (user User) {
	u.Model.FindByID(id, &user)
	return
}

func (u *UserModel) Store(users []User) *mongo.BulkWriteResult {
	if len(users) == 0 {
		return nil
	}
	var models []mongo.WriteModel
	for _, user := range users {
		user.parseLocation()
		filter := bson.D{{"_id", user.ID()}}
		update := bson.D{{"$set", user}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}
	return database.BulkWrite(u.Name(), models)
}

func (u *UserModel) UpdateGists(user User, gists []query.Gist) {
	filter := bson.D{{"_id", user.ID()}}
	update := bson.D{{"$set", bson.D{{"gists", gists}}}}
	database.UpdateOne(u.Name(), filter, update)
}

func (u *UserModel) UpdateRepositories(user User, repositories []Repository) {
	filter := bson.D{{"_id", user.ID()}}
	update := bson.D{{"$set", bson.D{{"repositories", repositories}}}}
	database.UpdateOne(u.Name(), filter, update)
}

func NewUserModel() *UserModel {
	return &UserModel{
		Model: &Model{
			name: "users",
		},
	}
}
