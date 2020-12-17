package app

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
)

type Rank struct {
	Rank       int       `bson:"rank"`
	TotalCount int       `bson:"total_count"`
	CreatedAt  time.Time `bson:"created_at"`
}

func Languages() (languages []string) {
	b, err := ioutil.ReadFile("./assets/languages.json")
	if err != nil {
		log.Fatalln(err.Error())
	}
	if err = json.Unmarshal(b, &languages); err != nil {
		log.Fatalln(err.Error())
	}
	return languages
}
