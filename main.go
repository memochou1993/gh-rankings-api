package main

import (
	"github.com/memochou1993/github-rankings/app/model"
	"log"
)

func main() {
	userCollection := model.UserCollection{}
	if err := userCollection.Init(); err != nil {
		log.Println(err.Error())
	}
	if err := userCollection.Collect(); err != nil {
		log.Println(err.Error())
	}
	if err := userCollection.Update(); err != nil {
		log.Println(err.Error())
	}
}
