package handler

import (
	"github.com/memochou1993/github-rankings/logger"
	"time"
)

type Handler struct {
	*RepositoryHandler
	*OwnerHandler
}

func NewHandler() *Handler {
	return &Handler{
		RepositoryHandler: NewRepositoryHandler(),
		OwnerHandler:      NewOwnerHandler(),
	}
}

func (h *Handler) Init() {
	h.RepositoryHandler.Init()
	h.OwnerHandler.Init()
	h.work()
}

func (h *Handler) work() {
	go h.collectRepositories()
	go h.rankRepositories()
	go h.collectOwners()
	go h.updateOwners()
	go h.rankOwners()
}

func (h *Handler) collectRepositories() {
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		if err := h.RepositoryHandler.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (h *Handler) rankRepositories() {
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		h.RepositoryHandler.Rank()
	}
}

func (h *Handler) collectOwners() {
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		if err := h.OwnerHandler.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (h *Handler) updateOwners() {
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		if err := h.OwnerHandler.Update(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (h *Handler) rankOwners() {
	t := time.NewTicker(10 * time.Minute)
	for ; true; <-t.C {
		h.OwnerHandler.Rank()
	}
}
