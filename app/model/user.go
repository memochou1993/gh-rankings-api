package model

import (
	"github.com/memochou1993/gh-rankings/app/query"
	"github.com/memochou1993/gh-rankings/app/resource"
	"github.com/memochou1993/gh-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type User struct {
	ImageUrl       string       `json:"imageUrl,omitempty" bson:"image_url"`
	CreatedAt      *time.Time   `json:"createdAt,omitempty" bson:"created_at"`
	Followers      *query.Items `json:"followers,omitempty" bson:"followers"`
	Location       string       `json:"location,omitempty" bson:"location"`
	Login          string       `json:"login" bson:"_id"`
	Name           string       `json:"name,omitempty" bson:"name"`
	Gists          []Gist       `json:"gists,omitempty" bson:"gists,omitempty"`
	Repositories   []Repository `json:"repositories,omitempty" bson:"repositories,omitempty"`
	ParsedLocation string       `json:"parsedLocation,omitempty" bson:"parsed_location"`
	ParsedCity     string       `json:"parsedCity,omitempty" bson:"parsed_city"`
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

func (u *UserModel) UpdateGists(user User, gists []Gist) {
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
		Model: NewModel("users"),
	}
}
