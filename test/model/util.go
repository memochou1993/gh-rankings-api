package model

import (
	"context"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/database"
	"log"
)

func dropDatabase() {
	if err := database.GetDatabase().Drop(context.Background()); err != nil {
		log.Fatalln(err.Error())
	}
}

func dropCollection(c model.CollectionInterface) {
	if err := c.GetCollection().Drop(context.Background()); err != nil {
		log.Fatalln(err.Error())
	}
}
