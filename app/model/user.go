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
	Gists        []Gist       `json:"gists" bson:"gists,omitempty"`
	Repositories []Repository `json:"repositories" bson:"repositories,omitempty"`
	Ranks        *struct {
		GistStars       *Rank `json:"gistStars" bson:"gist_stars,omitempty"`
		RepositoryStars *Rank `json:"repositoryStars" bson:"repository_stars,omitempty"`
	} `json:"ranks" bson:"ranks,omitempty"`
}

type UserResponse struct {
	Data struct {
		Search struct {
			Edges []struct {
				Cursor string `json:"cursor"`
				Node   User   `json:"node"`
			} `json:"edges"`
			PageInfo `json:"pageInfo"`
		} `json:"search"`
		User struct {
			AvatarURL string    `json:"avatarUrl"`
			CreatedAt time.Time `json:"createdAt"`
			Followers Directory `json:"followers"`
			Gists     struct {
				Edges []struct {
					Cursor string `json:"cursor"`
					Node   Gist   `json:"node"`
				} `json:"edges"`
				PageInfo   `json:"pageInfo"`
				TotalCount int `json:"totalCount"`
			} `json:"gists"`
			Location     string `json:"location"`
			Login        string `json:"login"`
			Name         string `json:"name"`
			Repositories struct {
				Edges []struct {
					Cursor string     `json:"cursor"`
					Node   Repository `json:"node"`
				} `json:"edges"`
				PageInfo   `json:"pageInfo"`
				TotalCount int `json:"totalCount"`
			} `json:"repositories"`
		} `json:"user"`
		RateLimit `json:"rateLimit"`
	} `json:"data"`
	Errors []Error `json:"errors"`
}

type UserRank struct {
	Login      string `bson:"_id"`
	TotalCount int    `bson:"total_count"`
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
