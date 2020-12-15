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

func (w *Worker) BuildUserCollection() {
	u := NewUserCollection()
	u.Init(w.starter)
	<-w.starter
	go w.collectUsers()
	go w.updateUsers()
	go w.rankUserRepositoryStars()
}

func (w *Worker) collectUsers() {
	u := NewUserCollection()
	t := time.NewTicker(24 * time.Hour)
	for ; true; <-t.C {
		if err := u.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (w *Worker) updateUsers() {
	u := NewUserCollection()
	t := time.NewTicker(24 * time.Hour)
	for ; true; <-t.C {
		if err := u.Update(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (w *Worker) rankUserRepositoryStars() {
	u := NewUserCollection()
	t := time.NewTicker(24 * time.Hour)
	for ; true; <-t.C {
		u.RankRepositoryStars()
	}
}
