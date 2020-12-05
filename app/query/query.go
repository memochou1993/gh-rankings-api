package query

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

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

	return strings.Replace(string(data), "<args>", join(args), 1)
}

func join(v interface{}) string {
	data, err := json.Marshal(v)
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
