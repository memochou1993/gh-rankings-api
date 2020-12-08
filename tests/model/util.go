package model

import (
	"context"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/database"
	"log"
)

func dropDatabase() {
	if err := database.GetDatabase().Drop(context.Background()); err != nil {
		log.Fatal(err.Error())
	}
}

func dropCollection(collection model.CollectionInterface) {
	if err := collection.GetCollection().Drop(context.Background()); err != nil {
		log.Fatal(err.Error())
	}
}
