package model

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func joinArguments(v interface{}) string {
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
