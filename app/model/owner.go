package model

import (
	"time"
)

type Owner struct {
	AvatarURL    string       `json:"avatarUrl" bson:"avatar_url"`
	CreatedAt    time.Time    `json:"createdAt" bson:"created_at"`
	Followers    *Directory   `json:"followers" bson:"followers"`
	Location     string       `json:"location" bson:"location"`
	Login        string       `json:"login" bson:"_id"`
	Name         string       `json:"name" bson:"name"`
	Gists        []Gist       `json:"gists" bson:"gists,omitempty"`
	Repositories []Repository `json:"repositories" bson:"repositories,omitempty"`
	Ranks        []Rank       `json:"ranks" bson:"ranks,omitempty"`
	Type         string       `json:"type" bson:"type"`
}

type OwnerResponse struct {
	Data struct {
		Search struct {
			Edges []struct {
				Cursor string `json:"cursor"`
				Node   Owner  `json:"node"`
			} `json:"edges"`
			PageInfo `json:"pageInfo"`
		} `json:"search"`
		Owner struct {
			AvatarURL string    `json:"avatarUrl"`
			CreatedAt time.Time `json:"createdAt"`
			Followers Directory `json:"followers"`
			Gists     struct {
				Edges []struct {
					Cursor string `json:"cursor"`
					Node   Gist   `json:"node"`
				} `json:"edges"`
				PageInfo `json:"pageInfo"`
			} `json:"gists"`
			Location     string `json:"location"`
			Login        string `json:"login"`
			Name         string `json:"name"`
			Repositories struct {
				Edges []struct {
					Cursor string     `json:"cursor"`
					Node   Repository `json:"node"`
				} `json:"edges"`
				PageInfo `json:"pageInfo"`
			} `json:"repositories"`
		} `json:"owner"`
		RateLimit `json:"rateLimit"`
	} `json:"data"`
	Errors []Error `json:"errors"`
}

type OwnerRank struct {
	Login      string `bson:"_id"`
	TotalCount int    `bson:"total_count"`
}

type OwnerModel struct {
	*Model
}

func NewOwnerModel() *OwnerModel {
	return &OwnerModel{
		&Model{
			name: "owners",
		},
	}
}
