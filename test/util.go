package test

import (
	"context"
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/database"
	"log"
	"os"
	"path"
	"runtime"
)

func ChangeDirectory() {
	_, file, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(file), "..")
	if err := os.Chdir(dir); err != nil {
		log.Fatalln(err.Error())
	}
}

func DropCollection(c app.CollectionInterface) {
	if err := c.GetCollection().Drop(context.Background()); err != nil {
		log.Fatalln(err.Error())
	}
}

func DropDatabase() {
	if err := database.GetDatabase().Drop(context.Background()); err != nil {
		log.Fatalln(err.Error())
	}
}
