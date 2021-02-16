package worker

import (
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/logger"
	"github.com/spf13/viper"
	"log"
	"time"
)

const (
	TimestampUser         = "TIMESTAMP_USER"
	TimestampOrganization = "TIMESTAMP_ORGANIZATION"
	TimestampRepository   = "TIMESTAMP_REPOSITORY"
)

var (
	UserWorker         = NewUserWorker()
	OrganizationWorker = NewOrganizationWorker()
	RepositoryWorker   = NewRepositoryWorker()
)

var (
	collecting int64
)

type Interface interface {
	Init()
	Collect() error
	Rank()
}

type Worker struct {
	Timestamp time.Time
}

func (w *Worker) load(timestamp string) {
	if timestamp := viper.GetInt64(timestamp); timestamp > 0 {
		w.Timestamp = time.Unix(0, timestamp)
	}
}

func (w *Worker) save(key string, t time.Time) {
	w.Timestamp = t
	viper.Set(key, t.UnixNano())
	if err := viper.WriteConfig(); err != nil {
		log.Fatal(err.Error())
	}
}

func Start() {
	model.NewRankModel().CreateIndexes()

	go run(UserWorker)
	go run(OrganizationWorker)
	go run(RepositoryWorker)
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
