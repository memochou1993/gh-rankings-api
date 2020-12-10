package main

import (
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"log"
)

func init() {
	util.LoadEnv()
}

func main() {
	database.Init()
	logger.Init()

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
