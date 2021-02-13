package test

import (
	"context"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/database"
	"log"
)

func DropCollection(c model.Interface) {
	if err := c.Collection().Drop(context.Background()); err != nil {
		log.Fatal(err.Error())
	}
}

func DropDatabase() {
	if err := database.Database().Drop(context.Background()); err != nil {
		log.Fatal(err.Error())
	}
}
