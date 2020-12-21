package model

import (
	"encoding/json"
	"fmt"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type Payload struct {
	Query string `json:"query"`
}

type Query struct {
	Schema string
	Field  string
	SearchArguments
	OwnerArguments
	GistsArguments
	RepositoriesArguments
}

func (q Query) String() string {
	query := q.Schema
	query = strings.Replace(query, "<Field>", q.Field, 1)
	query = strings.Replace(query, "<SearchArguments>", util.ParseStruct(q.SearchArguments, ","), 1)
	query = strings.Replace(query, "<OwnerArguments>", util.ParseStruct(q.OwnerArguments, ","), 1)
	query = strings.Replace(query, "<GistsArguments>", util.ParseStruct(q.GistsArguments, ","), 1)
	query = strings.Replace(query, "<RepositoriesArguments>", util.ParseStruct(q.RepositoriesArguments, ","), 1)

	b, err := json.Marshal(Payload{Query: query})
	if err != nil {
		log.Fatalln(err.Error())
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
	Repos     string `json:"repos,omitempty"`
	Sort      string `json:"sort,omitempty"`
	Type      string `json:"type,omitempty"`
}

type Directory struct {
	TotalCount int `json:"totalCount" bson:"total_count"`
}

type PageInfo struct {
	EndCursor   string `json:"endCursor,omitempty"`
	HasNextPage bool   `json:"hasNextPage,omitempty"`
}

type RateLimit struct {
	Cost      int    `json:"cost,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	NodeCount int    `json:"nodeCount,omitempty"`
	Remaining int    `json:"remaining,omitempty"`
	ResetAt   string `json:"resetAt,omitempty"`
	Used      int    `json:"used,omitempty"`
}

func (r RateLimit) Break() {
	buffer := 10
	if r.Remaining > buffer {
		return
	}
	resetAt, err := time.Parse(time.RFC3339, r.ResetAt)
	if err != nil {
		log.Fatalln(err.Error())
	}
	logger.Warning("Take a break...")
	time.Sleep(resetAt.Add(time.Second).Sub(time.Now().UTC()))
}

type Error struct {
	Type      string `json:"type"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

func NewOwnersQuery() *Query {
	return &Query{
		Schema: ReadQuery("owners"),
		SearchArguments: SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
}

func NewOwnerGistsQuery() *Query {
	return &Query{
		Schema: ReadQuery("owner_gists"),
		GistsArguments: GistsArguments{
			First:   100,
			OrderBy: "{field:CREATED_AT,direction:ASC}",
		},
	}
}

func NewOwnerRepositoriesQuery() *Query {
	return &Query{
		Schema: ReadQuery("owner_repositories"),
		RepositoriesArguments: RepositoriesArguments{
			First:             100,
			OrderBy:           "{field:CREATED_AT,direction:ASC}",
			OwnerAffiliations: "OWNER",
		},
	}
}

func ReadQuery(filename string) string {
	b, err := ioutil.ReadFile(fmt.Sprintf("./query/%s.graphql", filename))
	if err != nil {
		log.Fatalln(err.Error())
	}

	return string(b)
}