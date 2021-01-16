package worker

import (
	"github.com/memochou1993/github-rankings/logger"
	"time"
)

var (
	OwnerWorker      = NewOwnerWorker()
	RepositoryWorker = NewRepositoryWorker()
)

type Interface interface {
	Collect() error
	Rank()
}

func Init() {
	OwnerWorker.Init()
	go Collect(OwnerWorker)
	go Rank(OwnerWorker)

	RepositoryWorker.Init()
	go Collect(RepositoryWorker)
	go Rank(RepositoryWorker)
}

func Collect(worker Interface) {
	t := time.NewTicker(time.Hour)
	for ; true; <-t.C {
		if err := worker.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func Rank(worker Interface) {
	t := time.NewTicker(time.Hour)
	for ; true; <-t.C {
		worker.Rank()
	}
}
