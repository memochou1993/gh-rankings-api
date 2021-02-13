package query

import (
	"encoding/json"
	"fmt"
	"github.com/memochou1993/gh-rankings/util"
	"io/ioutil"
	"log"
	"strings"
)

type Query struct {
	Schema string
	Type   string
	*SearchArguments
	*OwnerArguments
	*GistsArguments
	*RepositoriesArguments
}

func (q Query) String() string {
	query := q.Schema
	query = strings.Replace(query, "<Type>", q.Type, 1)
	query = strings.Replace(query, "<SearchArguments>", util.ParseStruct(q.SearchArguments, ","), 1)
	query = strings.Replace(query, "<OwnerArguments>", util.ParseStruct(q.OwnerArguments, ","), 1)
	query = strings.Replace(query, "<GistsArguments>", util.ParseStruct(q.GistsArguments, ","), 1)
	query = strings.Replace(query, "<RepositoriesArguments>", util.ParseStruct(q.RepositoriesArguments, ","), 1)

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

type GistsArguments struct {
	After   string `json:"after,omitempty"`
	First   int    `json:"first,omitempty"`
	OrderBy string `json:"orderBy,omitempty"`
}

type RepositoriesArguments struct {
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

type Gist struct {
	Forks      *Items `json:"forks" bson:"forks"`
	Name       string `json:"name" bson:"name"`
	Stargazers *Items `json:"stargazers" bson:"stargazers"`
}

type Items struct {
	TotalCount int `json:"totalCount" bson:"total_count"`
}

func Owners() *Query {
	return &Query{
		Schema: read("owners"),
		SearchArguments: &SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
}

func OwnerGists() *Query {
	return &Query{
		Schema:         read("owner_gists"),
		OwnerArguments: &OwnerArguments{},
		GistsArguments: &GistsArguments{
			First:   100,
			OrderBy: "{field:CREATED_AT,direction:ASC}",
		},
	}
}

func OwnerRepositories() *Query {
	return &Query{
		Schema:         read("owner_repositories"),
		OwnerArguments: &OwnerArguments{},
		RepositoriesArguments: &RepositoriesArguments{
			First:             100,
			OrderBy:           "{field:CREATED_AT,direction:ASC}",
			OwnerAffiliations: "OWNER",
		},
	}
}

func Repositories() *Query {
	return &Query{
		Schema: read("repositories"),
		SearchArguments: &SearchArguments{
			First: 100,
			Type:  "REPOSITORY",
		},
	}
}

func read(name string) string {
	b, err := ioutil.ReadFile(fmt.Sprintf("%s/assets/query/%s.graphql", util.Root(), name))
	if err != nil {
		log.Fatal(err.Error())
	}
	return string(b)
}
