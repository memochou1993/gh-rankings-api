package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/memochou1993/github-rankings/app/model"
	"log"
)

func main() {
	userCollection := model.UserCollection{
		// TODO
	}
	if err := userCollection.Init(); err != nil {
		log.Println(err.Error())
	}
}
