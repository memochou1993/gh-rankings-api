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
	if rl.ResetAt == "" {
		return
	}
	resetAt, err := time.Parse(time.RFC3339, rl.ResetAt)
	if err != nil {
		log.Fatal(err.Error())
	}
	if rl.Remaining > 0 {
		return
	}
	duration := resetAt.Sub(time.Now().UTC())
	log.Println(fmt.Sprintf("Wait about %d minutes for next call", int(duration.Minutes())))
	time.Sleep(duration)
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
		log.Fatal(err.Error())
	}

	return strings.Replace(string(data), "<args>", util.JoinStruct(args), 1)
}
