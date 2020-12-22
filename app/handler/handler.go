package handler

import (
	"github.com/memochou1993/github-rankings/logger"
	"time"
)

type Handler struct {
	starter chan struct{}
	*RepositoryHandler
	*OwnerHandler
}

func NewHandler() *Handler {
	return &Handler{
		starter:           make(chan struct{}, 1),
		RepositoryHandler: NewRepositoryHandler(),
		OwnerHandler:      NewOwnerHandler(),
	}
}

func (h *Handler) Build() {
	h.RepositoryHandler.Init(h.starter)
	<-h.starter
	go h.collectRepositories()
	// TODO
	// go h.rankRepositories()
	h.OwnerHandler.Init(h.starter)
	<-h.starter
	go h.collectOwners()
	go h.updateOwners()
	go h.rankOwners()
}

func (h *Handler) collectRepositories() {
	t := time.NewTicker(10 * time.Minute) // FIXME
	for ; true; <-t.C {
		if err := h.RepositoryHandler.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

// func (h *Handler) rankRepositories() {
// 	t := time.NewTicker(10 * time.Minute) // FIXME
// 	for ; true; <-t.C {
// 		if err := h.RepositoryHandler.Rank(); err != nil {
// 			logger.Error(err.Error())
// 		}
// 	}
// }

func (h *Handler) collectOwners() {
	t := time.NewTicker(10 * time.Minute) // FIXME
	for ; true; <-t.C {
		if err := h.OwnerHandler.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (h *Handler) updateOwners() {
	t := time.NewTicker(10 * time.Minute) // FIXME
	for ; true; <-t.C {
		if err := h.OwnerHandler.Update(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (h *Handler) rankOwners() {
	t := time.NewTicker(10 * time.Minute) // FIXME
	for ; true; <-t.C {
		h.OwnerHandler.Rank()
	}
}
