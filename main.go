package main

import (
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"time"
)

func init() {
	util.LoadEnv()
	database.Init()
	logger.Init()
}

func main() {
	userCollection := model.UserCollection{}
	if err := userCollection.Init(); err != nil {
		logger.Error(err.Error())
	}
	go func() {
		t := time.NewTicker(10 * time.Second)
		for ; true; <-t.C {
			if err := userCollection.Collect(); err != nil {
				logger.Error(err.Error())
			}
		}
	}()
	// TODO
	// go func() {
	// 	time.Sleep(time.Second)
	// 	for {
	// 		if err := userCollection.Update(); err != nil {
	// 			logger.Error(err.Error())
	// 		}
	// 	}
	// }()

	// FIXME
	time.Sleep(time.Hour)
}
