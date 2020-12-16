package main

import (
	"github.com/memochou1993/github-rankings/app"
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
	work()

	time.Sleep(6 * time.Hour) // FIXME
}

func work() {
	worker := app.NewWorker()
	worker.BuildUserModel()
}
