package worker

import (
	"github.com/memochou1993/github-rankings/logger"
	"time"
)

var (
	OwnerWorker      = NewOwnerWorker()
	RepositoryWorker = NewRepositoryWorker()
)

func Init() {
	OwnerWorker.Init()
	go collectOwners()
	go rankOwners()

	RepositoryWorker.Init()
	go collectRepositories()
	go rankRepositories()
}

func collectOwners() {
	t := time.NewTicker(10 * time.Second)
	for ; true; <-t.C {
		if err := OwnerWorker.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func rankOwners() {
	t := time.NewTicker(24 * time.Hour)
	for ; true; <-t.C {
		OwnerWorker.Rank()
	}
}

func collectRepositories() {
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		if err := RepositoryWorker.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func rankRepositories() {
	t := time.NewTicker(24 * time.Hour)
	for ; true; <-t.C {
		RepositoryWorker.Rank()
	}
}
