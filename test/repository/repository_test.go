package repository

import (
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/worker"
	"github.com/memochou1993/gh-rankings/database"
	"github.com/memochou1993/gh-rankings/logger"
	"github.com/memochou1993/gh-rankings/test"
	"github.com/memochou1993/gh-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"strconv"
	"testing"
)

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	test.ChangeDirectory()
	util.LoadEnv()
	database.Init()
	logger.Init()
}

func TestFetch(t *testing.T) {
	r := worker.NewRepositoryWorker()

	r.SearchQuery = model.NewRepositoryQuery()
	r.SearchQuery.SearchArguments.Query = strconv.Quote("created:2020-01-01..2020-01-01 fork:true sort:stars stars:>=100")

	var repositories []model.Repository
	if err := r.Fetch(&repositories); err != nil {
		t.Error(err.Error())
	}
	if len(repositories) == 0 {
		t.Fail()
	}

	test.DropCollection(r.RepositoryModel)
}

func TestStore(t *testing.T) {
	r := worker.NewRepositoryWorker()

	repository := model.Repository{NameWithOwner: "memochou1993/gh-rankings"}
	repositories := []model.Repository{repository}

	r.RepositoryModel.Store(repositories)
	res := database.FindOne(r.RepositoryModel.Name(), bson.D{{"_id", repository.ID()}})
	if res.Err() == mongo.ErrNoDocuments {
		t.Fail()
	}

	repository = model.Repository{}
	if err := res.Decode(&repository); err != nil {
		t.Fatal(err.Error())
	}

	test.DropCollection(r.RepositoryModel)
}

func tearDown() {
	test.DropDatabase()
}
