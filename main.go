package main

import (
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"log"

	_ "github.com/joho/godotenv/autoload"
)

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
