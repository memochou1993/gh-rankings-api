package handler

import (
	"github.com/memochou1993/github-rankings/logger"
	"time"
)

type Interface interface {
	Init(starter chan<- struct{})
	Collect() error
	Update() error
	Rank()
}

type Handler struct {
	starter chan struct{}
}

func NewHandler() *Handler {
	return &Handler{
		starter: make(chan struct{}, 1),
	}
}

func (h *Handler) Build(handler Interface) {
	handler.Init(h.starter)
	<-h.starter
	go h.collect(handler)
	// go h.update(handler)
	// go h.rank(handler)
}

func (h *Handler) collect(handler Interface) {
	t := time.NewTicker(10 * time.Minute) // FIXME
	for ; true; <-t.C {
		if err := handler.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (h *Handler) update(handler Interface) {
	t := time.NewTicker(10 * time.Minute) // FIXME
	for ; true; <-t.C {
		if err := handler.Update(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (h *Handler) rank(handler Interface) {
	t := time.NewTicker(10 * time.Minute) // FIXME
	for ; true; <-t.C {
		handler.Rank()
	}
}
