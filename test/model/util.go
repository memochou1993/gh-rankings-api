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
	_, file, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(file), "../..")
	if err := os.Chdir(dir); err != nil {
		log.Fatalln(err)
	}
}

func dropCollection(c model.CollectionInterface) {
	if err := c.GetCollection().Drop(context.Background()); err != nil {
		log.Fatalln(err.Error())
	}
}

func dropDatabase() {
	if err := database.GetDatabase().Drop(context.Background()); err != nil {
		log.Fatalln(err.Error())
	}
}
