package user

import (
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/query"
	"github.com/memochou1993/gh-rankings/app/worker"
	"github.com/memochou1993/gh-rankings/database"
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
	util.LoadEnv()
	database.Connect()
}

func TestFetch(t *testing.T) {
	u := worker.NewUserWorker()

	u.SearchQuery = query.Owners()
	u.SearchQuery.SearchArguments.Query = strconv.Quote("created:2020-01-01..2020-01-01 followers:100..* sort:joined-asc")

	var users []model.User
	if err := u.Fetch(&users); err != nil {
		t.Error(err.Error())
	}
	if len(users) == 0 {
		t.Fail()
	}

	test.DropCollection(u.UserModel)
}

func TestStore(t *testing.T) {
	u := worker.NewUserWorker()

	user := model.User{Login: "memochou1993", Followers: &query.Items{TotalCount: 1}}
	users := []model.User{user}

	u.UserModel.Store(users)
	res := database.FindOne(u.UserModel.Name(), bson.D{{"_id", user.ID()}})
	if res.Err() == mongo.ErrNoDocuments {
		t.Fail()
	}

	user = model.User{}
	if err := res.Decode(&user); err != nil {
		t.Fatal(err.Error())
	}

	test.DropCollection(u.UserModel)
}

func TestFetchGists(t *testing.T) {
	u := worker.NewUserWorker()

	u.GistQuery = query.OwnerGists()
	u.GistQuery.Type = model.TypeUser
	u.GistQuery.OwnerArguments.Login = strconv.Quote("memochou1993")

	var gists []query.Gist
	if err := u.FetchGists(&gists); err != nil {
		t.Error(err.Error())
	}
	if len(gists) == 0 {
		t.Fail()
	}

	test.DropCollection(u.UserModel)
}

func TestFetchRepositories(t *testing.T) {
	u := worker.NewUserWorker()

	u.RepositoryQuery = query.OwnerRepositories()
	u.RepositoryQuery.Type = model.TypeUser
	u.RepositoryQuery.OwnerArguments.Login = strconv.Quote("memochou1993")

	var repositories []model.Repository
	if err := u.FetchRepositories(&repositories); err != nil {
		t.Error(err.Error())
	}
	if len(repositories) == 0 {
		t.Fail()
	}

	test.DropCollection(u.UserModel)
}

func tearDown() {
	test.DropDatabase()
}
