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

func (w *Worker) BuildUserCollection() {
	w.userCollection.Init(w.starter)
	<-w.starter

	go func() {
		for range time.Tick(time.Second) {
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
