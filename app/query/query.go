package query

import (
	"fmt"
	"github.com/memochou1993/github-rankings/util"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

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

type SearchArguments struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
	First  int    `json:"first,omitempty"`
	Last   int    `json:"last,omitempty"`
	Query  string `json:"query,omitempty"`
	Type   string `json:"type,omitempty"`
}

func (args *SearchArguments) Read(query string) string {
	data, err := ioutil.ReadFile(fmt.Sprintf("./app/query/%s.graphql", query))
	if err != nil {
		log.Fatalln(err.Error())
	}

	return strings.Replace(string(data), "<args>", util.JoinStruct(args), 1)
}

type Error struct {
	Type      string `json:"type"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Message string `json:"message"`
}
