package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/memochou1993/github-rankings/app/database"
	"log"
)

func main() {
	if err := database.CollectUsers(); err != nil {
		log.Println(err.Error())
	}
}
