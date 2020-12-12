package app

import (
	"github.com/memochou1993/github-rankings/logger"
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

func (w *Worker) CollectUsers() {
	if err := w.userCollection.Init(w.starter); err != nil {
		logger.Error(err.Error())
	}

	go func() {
		<-w.starter
		close(w.starter)
		t := time.NewTicker(7 * 24 * time.Hour)
		for ; true; <-t.C {
			if err := w.userCollection.Collect(); err != nil {
				logger.Error(err.Error())
				time.Sleep(time.Hour)
			}
		}
	}()

	go func() {
		for range time.Tick(time.Second) {
			if err := w.userCollection.Update(); err != nil {
				logger.Error(err.Error())
				time.Sleep(time.Hour)
			}
		}
	}()
}
