package model

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

const CollectionUsers = "users"

type SearchArguments struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
	First  int    `json:"first,omitempty"`
	Last   int    `json:"last,omitempty"`
	Query  string `json:"query,omitempty"`
	Type   string `json:"type,omitempty"`
}

type Users struct {
	Data struct {
		Search struct {
			UserCount int `json:"userCount" bson:"userCount"`
			Edges     []struct {
				Cursor string `json:"cursor"`
				Node   struct {
					ID    string `json:"id"`
					Login string `json:"login"`
				} `json:"node"`
			} `json:"edges"`
			PageInfo struct {
				EndCursor   string `json:"endCursor"`
				HasNextPage bool   `json:"hasNextPage"`
				StartCursor string `json:"startCursor"`
			} `json:"pageInfo"`
		} `json:"search"`
	} `json:"data"`
}

func (u *Users) GetQuery(args SearchArguments) string {
	data, err := ioutil.ReadFile(fmt.Sprintf("./app/model/%s.graphql", CollectionUsers))
	if err != nil {
		log.Fatal(err.Error())
	}

	return strings.Replace(string(data), "<args>", joinArguments(&args), 1)
}
