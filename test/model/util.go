package model

import (
	"context"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/database"
	"log"
	"os"
	"path"
	"runtime"
)

func changeDirectory() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	if err := os.Chdir(dir); err != nil {
		log.Fatalln(err)
	}
}

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
