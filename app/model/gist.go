package model

import "github.com/memochou1993/gh-rankings/app/query"

type Gist struct {
	Forks      *query.Items `json:"forks" bson:"forks"`
	Name       string       `json:"name" bson:"name"`
	Stargazers *query.Items `json:"stargazers" bson:"stargazers"`
}
