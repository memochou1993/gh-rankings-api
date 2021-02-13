package worker

import (
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/logger"
	"github.com/spf13/viper"
	"log"
	"time"
)

const (
	TimestampUserRanks         = "TIMESTAMP_USER_RANKS"
	TimestampOrganizationRanks = "TIMESTAMP_ORGANIZATION_RANKS"
	TimestampRepositoryRanks   = "TIMESTAMP_REPOSITORY_RANKS"
)

var (
	userWorker         = NewUserWorker()
	organizationWorker = NewOrganizationWorker()
	repositoryWorker   = NewRepositoryWorker()
)

var (
	collecting int
)

type Interface interface {
	Init()
	Collect() error
	Rank()
}

type Worker struct {
	Timestamp time.Time
}

func (w *Worker) seal(key string, t time.Time) {
	w.Timestamp = t
	viper.Set(key, t.UnixNano())
	if err := viper.WriteConfig(); err != nil {
		log.Fatal(err.Error())
	}
}

func Start() {
	go run(userWorker)
	go run(organizationWorker)
	go run(repositoryWorker)
}

func List(req *request.Request) (ranks []model.Rank) {
	switch req.Type {
	case model.TypeUser:
		ranks = userWorker.List(req)
	case model.TypeOrganization:
		ranks = organizationWorker.List(req)
	case model.TypeRepository:
		ranks = repositoryWorker.List(req)
	}
	return
}

func run(worker Interface) {
	worker.Init()
	t := time.NewTicker(24 * time.Hour)
	for ; true; <-t.C {
		collecting += 1
		if err := worker.Collect(); err != nil {
			logger.Error(err.Error())
		}
		collecting -= 1
		worker.Rank()
	}
}
