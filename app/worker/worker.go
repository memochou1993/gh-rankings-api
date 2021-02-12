package worker

import (
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
	collecting int
)

var (
	UserWorker         *userWorker
	OrganizationWorker *organizationWorker
	RepositoryWorker   *repositoryWorker
)

type Interface interface {
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

func Init() {
	// FIXME: create indexes
	// RankModel.CreateIndexes()

	UserWorker = NewUserWorker()
	go Run(UserWorker)

	OrganizationWorker = NewOrganizationWorker()
	go Run(OrganizationWorker)

	RepositoryWorker = NewRepositoryWorker()
	go Run(RepositoryWorker)
}

func Run(worker Interface) {
	t := time.NewTicker(24 * time.Hour)
	for ; true; <-t.C {
		// collecting += 1
		// if err := worker.Collect(); err != nil {
		// 	logger.Error(err.Error())
		// }
		// collecting -= 1
		worker.Rank()
	}
}

func NewWorker() *Worker {
	return &Worker{}
}
