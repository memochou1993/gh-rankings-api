package main

import (
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"time"
)

var (
	userCollection *model.UserCollection
)

func init() {
	util.LoadEnv()
	database.Init()
	logger.Init()
}

func main() {
	starter := make(chan struct{}, 1)

	userCollection = model.NewUserCollection()
	if err := userCollection.Init(starter); err != nil {
		logger.Error(err.Error())
	}

	go func() {
		<-starter
		close(starter)
		t := time.NewTicker(7 * 24 * time.Hour)
		for ; true; <-t.C {
			if err := userCollection.Collect(); err != nil {
				logger.Error(err.Error())
			}
		}
	}()

	go func() {
		for range time.Tick(time.Second) {
			if err := userCollection.Update(); err != nil {
				logger.Error(err.Error())
			}
		}
	}()

	// FIXME
	time.Sleep(time.Hour)
}
