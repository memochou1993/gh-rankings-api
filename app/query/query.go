package query

import (
	"fmt"
	"github.com/memochou1993/github-rankings/util"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type Request struct {
	Schema                []byte
	UserArguments         UserArguments
	SearchArguments       SearchArguments
	RepositoriesArguments RepositoriesArguments
}

func (r *Request) Join() []byte {
	s := string(r.Schema)
	s = strings.Replace(s, "UserArguments", util.JoinStruct(r.UserArguments, ","), 1)
	s = strings.Replace(s, "SearchArguments", util.JoinStruct(r.SearchArguments, ","), 1)
	s = strings.Replace(s, "RepositoriesArguments", util.JoinStruct(r.RepositoriesArguments, ","), 1)

	return []byte(s)
}

type UserArguments struct {
	Login string `json:"login,omitempty"`
}

type SearchArguments struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
	First  int    `json:"first,omitempty"`
	Last   int    `json:"last,omitempty"`
	Query  string `json:"query,omitempty"`
	Type   string `json:"type,omitempty"`
}

type RepositoriesArguments struct {
	After             string `json:"after,omitempty"`
	Before            string `json:"before,omitempty"`
	First             int    `json:"first,omitempty"`
	Last              int    `json:"last,omitempty"`
	OrderBy           string `json:"orderBy,omitempty"`
	OwnerAffiliations string `json:"ownerAffiliations,omitempty"`
}

type ArgumentsQuery struct {
	Created   string `json:"created,omitempty"`
	Followers string `json:"followers,omitempty"`
	Repos     string `json:"repos,omitempty"`
}

func (q *ArgumentsQuery) Join() string {
	return fmt.Sprintf("\"%s\"", util.JoinStruct(q, " "))
}

type PageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
	StartCursor string `json:"startCursor"`
}

type RateLimit struct {
	Cost      int
	Limit     int
	NodeCount int
	Remaining int
	ResetAt   string
	Used      int
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

func Read(filename string) []byte {
	data, err := ioutil.ReadFile(fmt.Sprintf("./app/query/%s.graphql", filename))
	if err != nil {
		log.Fatalln(err.Error())
	}

	return data
}
