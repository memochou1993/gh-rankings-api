package model

import (
	"time"
)

type Organization struct {
	AvatarURL    string       `json:"avatarUrl" bson:"avatar_url"`
	CreatedAt    time.Time    `json:"createdAt" bson:"created_at"`
	Location     string       `json:"location" bson:"location"`
	Login        string       `json:"login" bson:"_id"`
	Name         string       `json:"name" bson:"name"`
	Repositories []Repository `json:"repositories" bson:"repositories,omitempty"`
	Ranks        *struct {
		RepositoryStars *Rank `json:"repositoryStars" bson:"repository_stars,omitempty"`
	} `json:"ranks" bson:"ranks,omitempty"`
}

type OrganizationResponse struct {
	Data struct {
		Search struct {
			Edges []struct {
				Cursor string       `json:"cursor"`
				Node   Organization `json:"node"`
			} `json:"edges"`
			PageInfo `json:"pageInfo"`
		} `json:"search"`
		Organization struct {
			AvatarURL    string    `json:"avatarUrl"`
			CreatedAt    time.Time `json:"createdAt"`
			Location     string    `json:"location"`
			Login        string    `json:"login"`
			Name         string    `json:"name"`
			Repositories struct {
				Edges []struct {
					Cursor string     `json:"cursor"`
					Node   Repository `json:"node"`
				} `json:"edges"`
				PageInfo `json:"pageInfo"`
			} `json:"repositories"`
		} `json:"organization"`
		RateLimit `json:"rateLimit"`
	} `json:"data"`
	Errors []Error `json:"errors"`
}

type OrganizationRank struct {
	Login      string `bson:"_id"`
	TotalCount int    `bson:"total_count"`
}

type OrganizationModel struct {
	*Model
}

func NewOrganizationModel() *OrganizationModel {
	return &OrganizationModel{
		&Model{
			name: "organizations",
		},
	}
}
