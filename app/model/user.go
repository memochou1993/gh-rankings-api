package model

import (
	"github.com/memochou1993/gh-rankings/app/resource"
	"github.com/memochou1993/gh-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type User struct {
	AvatarURL      string       `json:"avatarUrl,omitempty" bson:"avatar_url,omitempty"`
	CreatedAt      *time.Time   `json:"createdAt,omitempty" bson:"created_at,omitempty"`
	Followers      *Items       `json:"followers,omitempty" bson:"followers,omitempty"`
	Location       string       `json:"location,omitempty" bson:"location,omitempty"`
	Login          string       `json:"login" bson:"_id"`
	Name           string       `json:"name,omitempty" bson:"name,omitempty"`
	Gists          []Gist       `json:"gists,omitempty" bson:"gists,omitempty"`
	Repositories   []Repository `json:"repositories,omitempty" bson:"repositories,omitempty"`
	ParsedLocation string       `json:"parsedLocation,omitempty" bson:"parsed_location,omitempty"`
	ParsedCity     string       `json:"parsedCity,omitempty" bson:"parsed_city,omitempty"`
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

func (u *UserModel) FindLast() (user User) {
	opts := options.FindOne().SetSort(bson.D{{"$natural", -1}})
	res := database.FindOne(u.Name(), bson.D{}, opts)
	if err := res.Decode(&user); err != nil && err != mongo.ErrNoDocuments {
		log.Fatal(err.Error())
	}
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
		&Model{
			name: "users",
		},
	}
}
