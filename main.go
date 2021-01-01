package main

import (
	"github.com/memochou1993/github-rankings/app/worker"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"time"
)

func init() {
	util.LoadEnv()
	database.Init()
	logger.Init()
	worker.Init()
}

func main() {
	w := worker.NewWorker()
	w.Init()

	time.Sleep(6 * time.Hour) // FIXME
}
