package model

import (
	"time"
)

type User struct {
	AvatarURL    string       `json:"avatarUrl" bson:"avatar_url"`
	CreatedAt    time.Time    `json:"createdAt" bson:"created_at"`
	Followers    Directory    `json:"followers" bson:"followers"`
	Location     string       `json:"location" bson:"location"`
	Login        string       `json:"login" bson:"_id"`
	Name         string       `json:"name" bson:"name"`
	Repositories []Repository `json:"repositories" bson:"repositories,omitempty"`
	Ranks        *struct {
		GistStars       *Rank `bson:"gist_stars,omitempty"`
		RepositoryStars *Rank `bson:"repository_stars,omitempty"`
	} `bson:"ranks,omitempty"`
}

type UserModel struct {
	*Model
}

func NewUserModel() *UserModel {
	return &UserModel{
		&Model{
			name: "users",
		},
	}
}
