package worker

import (
	"github.com/memochou1993/github-rankings/logger"
	"time"
)

var (
	Repository *RepositoryWorker
	Owner      *OwnerWorker
)

func Init() {
	Owner = NewOwnerWorker()
	Owner.Init()

	Repository = NewRepositoryWorker()
	Repository.Init()

	work()
}

func work() {
	go collectRepositories()
	go rankRepositories()
	go collectOwners()
	go updateOwners()
	go rankOwners()
}

func collectRepositories() {
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		if err := Repository.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func rankRepositories() {
	t := time.NewTicker(24 * time.Hour)
	for ; true; <-t.C {
		Repository.Rank()
	}
}

func collectOwners() {
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		if err := Owner.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func updateOwners() {
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		if err := Owner.Update(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func rankOwners() {
	t := time.NewTicker(24 * time.Hour)
	for ; true; <-t.C {
		Owner.Rank()
	}
}
