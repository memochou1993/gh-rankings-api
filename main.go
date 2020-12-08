package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/memochou1993/github-rankings/app/model"
	"log"
)

func main() {
	users := model.Users{}
	if err := users.Init(); err != nil {
		log.Println(err.Error())
	}
}
