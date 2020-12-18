package handler

import (
	"github.com/memochou1993/github-rankings/logger"
	"time"
)

type Handler struct {
	starter chan struct{}
}

func NewHandler() *Handler {
	return &Handler{
		starter: make(chan struct{}, 1),
	}
}

func (h *Handler) BuildUserModel() {
	u := NewUserHandler()
	u.Init(h.starter)
	<-h.starter
	go h.collectUsers()
	go h.updateUsers()
	go h.rankUsers()
}

func (h *Handler) collectUsers() {
	u := NewUserHandler()
	t := time.NewTicker(10 * time.Minute) // FIXME
	for ; true; <-t.C {
		if err := u.Collect(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (h *Handler) updateUsers() {
	u := NewUserHandler()
	t := time.NewTicker(10 * time.Minute) // FIXME
	for ; true; <-t.C {
		if err := u.Update(); err != nil {
			logger.Error(err.Error())
		}
	}
}

func (h *Handler) rankUsers() {
	u := NewUserHandler()
	t := time.NewTicker(10 * time.Minute) // FIXME
	for ; true; <-t.C {
		u.RankFollowers()
		u.RankGistStars()
		u.RankRepositoryStars()
		u.RankRepositoryStarsByLanguage()
	}
}
