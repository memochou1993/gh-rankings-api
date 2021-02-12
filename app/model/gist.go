package model

import "github.com/memochou1993/gh-rankings/app/query"

type Gist struct {
	Forks      *query.Items `json:"forks,omitempty" bson:"forks"`
	Name       string       `json:"name,omitempty" bson:"name"`
	Stargazers *query.Items `json:"stargazers,omitempty" bson:"stargazers"`
}
