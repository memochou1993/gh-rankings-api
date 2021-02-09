package test

import (
	"context"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/database"
	"log"
	"os"
	"path"
	"runtime"
)

func ChangeDirectory() {
	_, file, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(file), "../")
	if err := os.Chdir(dir); err != nil {
		log.Fatal(err.Error())
	}
}

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
