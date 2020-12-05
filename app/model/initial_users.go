package model

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type Arguments struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
	First  int    `json:"first,omitempty"`
	Last   int    `json:"last,omitempty"`
	Query  string `json:"query,omitempty"`
	Type   string `json:"type,omitempty"`
}

type InitialUsers struct {
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

func (u *InitialUsers) GetQuery(args Arguments) string {
	return fmt.Sprintf(`
		query InitialUsers {
		  search(%s) {
			userCount
			edges {
			  cursor
			  node {
				... on User {
				  id
				  login
				}
			  }
			}
			pageInfo {
			  endCursor
			  hasNextPage
			  startCursor
			}
		  }
		}`,
		args.join(),
	)
}

func (a *Arguments) join() string {
	data, err := json.Marshal(a)

	if err != nil {
		log.Fatal(err.Error())
	}

	var fields map[string]interface{}

	if err := json.Unmarshal(data, &fields); err != nil {
		log.Fatal(err.Error())
	}

	var args []string
	for field, value := range fields {
		args = append(args, fmt.Sprintf("%s: %v", field, value))
	}

	return strings.Join(args, ", ")
}
