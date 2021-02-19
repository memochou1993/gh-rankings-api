package worker

import (
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/logger"
	"github.com/spf13/viper"
	"log"
	"time"
)

const (
	timestampUser         = "TIMESTAMP_USER"
	timestampOrganization = "TIMESTAMP_ORGANIZATION"
	timestampRepository   = "TIMESTAMP_REPOSITORY"
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

	go run(UserWorker, 24*time.Hour)
	go run(OrganizationWorker, 24*time.Hour)
	go run(RepositoryWorker, 7*24*time.Hour)
}

func run(worker Interface, d time.Duration) {
	worker.Init()

	t := time.NewTicker(d)
	for ; true; <-t.C {
		var err error
		collecting += 1
		if err = worker.Collect(); err != nil {
			logger.Error(err.Error())
		}
		collecting -= 1
		if err != nil {
			return
		}
		worker.Rank()
	}
}
