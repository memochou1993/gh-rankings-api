package owner

import (
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/app/worker"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/test"
	"github.com/memochou1993/github-rankings/util"
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

func TestFetchRepositories(t *testing.T) {
	r := worker.RepositoryWorker

	q := model.Query{
		Schema: model.ReadQuery("search_repositories"),
		SearchArguments: model.SearchArguments{
			First: 100,
			Query: strconv.Quote("created:2020-01-01..2020-01-01 fork:true sort:stars stars:>=100"),
			Type:  "REPOSITORY",
		},
	}

	var repositories []model.Repository
	if err := r.FetchRepositories(&q, &repositories); err != nil {
		t.Error(err.Error())
	}
	if len(repositories) == 0 {
		t.Fail()
	}

	test.DropCollection(r.RepositoryModel)
}

func TestStoreRepositories(t *testing.T) {
	r := worker.RepositoryWorker

	repository := model.Repository{NameWithOwner: "memochou1993/github-rankings"}
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
