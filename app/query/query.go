package query

import (
	"encoding/json"
	"github.com/memochou1993/gh-rankings/util"
	"log"
	"strings"
)

type Query struct {
	Schema string
	Type   string
	SearchArguments
	OwnerArguments
	GistArguments
	RepositoryArguments
}

func (q Query) String() string {
	query := q.Schema
	query = strings.Replace(query, "<Type>", q.Type, 1)
	query = strings.Replace(query, "<SearchArguments>", util.ParseStruct(q.SearchArguments, ","), 1)
	query = strings.Replace(query, "<OwnerArguments>", util.ParseStruct(q.OwnerArguments, ","), 1)
	query = strings.Replace(query, "<GistArguments>", util.ParseStruct(q.GistArguments, ","), 1)
	query = strings.Replace(query, "<RepositoryArguments>", util.ParseStruct(q.RepositoryArguments, ","), 1)

	payload := struct {
		Query string `json:"query"`
	}{
		Query: query,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err.Error())
	}

	return string(b)
}

type SearchArguments struct {
	After string `json:"after,omitempty"`
	First int    `json:"first,omitempty"`
	Query string `json:"query,omitempty"`
	Type  string `json:"type,omitempty"`
}

type OwnerArguments struct {
	Login string `json:"login,omitempty"`
}

type GistArguments struct {
	After   string `json:"after,omitempty"`
	First   int    `json:"first,omitempty"`
	OrderBy string `json:"orderBy,omitempty"`
}

type RepositoryArguments struct {
	After             string `json:"after,omitempty"`
	First             int    `json:"first,omitempty"`
	OrderBy           string `json:"orderBy,omitempty"`
	OwnerAffiliations string `json:"ownerAffiliations,omitempty"`
}

type SearchQuery struct {
	Created   string `json:"created,omitempty"`
	Followers string `json:"followers,omitempty"`
	Fork      string `json:"fork,omitempty"`
	Repos     string `json:"repos,omitempty"`
	Sort      string `json:"sort,omitempty"`
	Stars     string `json:"stars,omitempty"`
	Type      string `json:"type,omitempty"`
}

type Items struct {
	TotalCount int `json:"totalCount,omitempty" bson:"total_count"`
}

func NewOwnerQuery() *Query {
	return &Query{
		Schema: util.ReadQuery("owners"),
		SearchArguments: SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
}

func NewOwnerGistQuery() *Query {
	return &Query{
		Schema: util.ReadQuery("owner_gists"),
		GistArguments: GistArguments{
			First:   100,
			OrderBy: "{field:CREATED_AT,direction:ASC}",
		},
	}
}

func NewOwnerRepositoryQuery() *Query {
	return &Query{
		Schema: util.ReadQuery("owner_repositories"),
		RepositoryArguments: RepositoryArguments{
			First:             100,
			OrderBy:           "{field:CREATED_AT,direction:ASC}",
			OwnerAffiliations: "OWNER",
		},
	}
}

func NewRepositoryQuery() *Query {
	return &Query{
		Schema: util.ReadQuery("repositories"),
		SearchArguments: SearchArguments{
			First: 100,
			Type:  "REPOSITORY",
		},
	}
}
