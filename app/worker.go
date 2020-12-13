package app

import (
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"time"
)

type Worker struct {
	starter        chan struct{}
	userCollection *UserCollection
}

func NewWorker() *Worker {
	return &Worker{
		starter:        make(chan struct{}, 1),
		userCollection: NewUserCollection(),
	}
}

func (w *Worker) BuildUserCollection() {
	w.userCollection.Init(w.starter)
	<-w.starter
	go w.CollectUsers()
	time.Sleep(10 * time.Second)
	go w.UpdateUsers()
	time.Sleep(10 * time.Second)
	go w.RankUserRepositoryStars()
}

func (w *Worker) CollectUsers() {
	duration := 7 * 24 * time.Hour
	if util.IsLocal() {
		duration = time.Second
	}

	t := time.NewTicker(duration)
	for ; true; <-t.C {
		if err := w.userCollection.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (w *Worker) UpdateUsers() {
	duration := 7 * 24 * time.Hour
	if util.IsLocal() {
		duration = time.Second
	}

	t := time.NewTicker(duration)
	for ; true; <-t.C {
		if err := w.userCollection.Update(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (w *Worker) RankUserRepositoryStars() {
	duration := 24 * time.Hour
	if util.IsLocal() {
		duration = 10 * time.Second
	}

	t := time.NewTicker(duration)
	for ; true; <-t.C {
		w.userCollection.RankRepositoryStars()
	}
}
