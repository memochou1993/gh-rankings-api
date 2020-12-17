package app

import (
	"github.com/memochou1993/github-rankings/logger"
	"time"
)

type Worker struct {
	starter chan struct{}
}

func NewWorker() *Worker {
	return &Worker{
		starter: make(chan struct{}, 1),
	}
}

func (w *Worker) BuildUserModel() {
	u := NewUserModel()
	u.Init(w.starter)
	<-w.starter
	go w.collectUsers()
	go w.updateUsers()
	go w.rankUsers()
}

func (w *Worker) collectUsers() {
	u := NewUserModel()
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		if err := u.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (w *Worker) updateUsers() {
	u := NewUserModel()
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		if err := u.Update(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (w *Worker) rankUsers() {
	u := NewUserModel()
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		u.RankGistStars()
		u.RankRepositoryStars()
	}
}
