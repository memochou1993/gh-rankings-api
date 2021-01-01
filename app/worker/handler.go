package worker

import (
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"time"
)

var (
	languages *model.Languages
	locations *model.Locations
)

type Worker struct {
	*RepositoryWorker
	*OwnerWorker
}

func init() {
	util.LoadAsset("languages", &languages)
	util.LoadAsset("locations", &locations)
}

func NewWorker() *Worker {
	return &Worker{
		RepositoryWorker: NewRepositoryWorker(),
		OwnerWorker:      NewOwnerWorker(),
	}
}

func (w *Worker) Init() {
	w.RepositoryWorker.Init()
	w.OwnerWorker.Init()
	w.work()
}

func (w *Worker) work() {
	go w.collectRepositories()
	go w.rankRepositories()
	go w.collectOwners()
	go w.updateOwners()
	go w.rankOwners()
}

func (w *Worker) collectRepositories() {
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		if err := w.RepositoryWorker.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (w *Worker) rankRepositories() {
	t := time.NewTicker(10 * time.Minute) // FIXME
	for ; true; <-t.C {
		w.RepositoryWorker.Rank()
	}
}

func (w *Worker) collectOwners() {
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		if err := w.OwnerWorker.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (w *Worker) updateOwners() {
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		if err := w.OwnerWorker.Update(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (w *Worker) rankOwners() {
	t := time.NewTicker(10 * time.Minute) // FIXME
	for ; true; <-t.C {
		w.OwnerWorker.Rank()
	}
}
