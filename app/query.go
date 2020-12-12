package app

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

type Request struct {
	Schema                string
	UserArguments         UserArguments
	SearchArguments       SearchArguments
	RepositoriesArguments RepositoriesArguments
}

func (r *Request) Query() string {
	q := r.Schema
	q = strings.Replace(q, "UserArguments", util.JoinStruct(r.UserArguments, ","), 1)
	q = strings.Replace(q, "SearchArguments", util.JoinStruct(r.SearchArguments, ","), 1)
	q = strings.Replace(q, "RepositoriesArguments", util.JoinStruct(r.RepositoriesArguments, ","), 1)

	query := struct {
		Query string `json:"query"`
	}{
		Query: q,
	}

	b, err := json.Marshal(query)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return string(b)
}

func (r *Request) Range(from string, to string) string {
	return fmt.Sprintf("%s..%s", from, to)
}

func (r *Request) String(v string) string {
	if v == "" {
		return v
	}
	return fmt.Sprintf("\"%s\"", v)
}

type UserArguments struct {
	Login string `json:"login,omitempty"`
}

type SearchArguments struct {
	After string `json:"after,omitempty"`
	First int    `json:"first,omitempty"`
	Query string `json:"query,omitempty"`
	Type  string `json:"type,omitempty"`
}

type RepositoriesArguments struct {
	After             string `json:"after,omitempty"`
	First             int    `json:"first,omitempty"`
	OrderBy           string `json:"orderBy,omitempty"`
	OwnerAffiliations string `json:"ownerAffiliations,omitempty"`
}

type ArgumentsQuery struct {
	Created   string `json:"created,omitempty"`
	Followers string `json:"followers,omitempty"`
	Repos     string `json:"repos,omitempty"`
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

func (rl *RateLimit) Check() {
	if rl.Remaining > 0 {
		return
	}
	if rl.ResetAt == "" {
		return
	}
	resetAt, err := time.Parse(time.RFC3339, rl.ResetAt)
	if err != nil {
		log.Fatalln(err.Error())
	}
	logger.Warning("Take a break...")
	time.Sleep(resetAt.Sub(time.Now().UTC()))
}

type Error struct {
	Type      string `json:"type"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Message string `json:"message"`
}

func ReadQuery(filename string) string {
	data, err := ioutil.ReadFile(fmt.Sprintf("./query/%s.graphql", filename))
	if err != nil {
		log.Fatalln(err.Error())
	}

	return string(data)
}
