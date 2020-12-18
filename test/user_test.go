package test

import (
	"github.com/memochou1993/github-rankings/app/handler"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	ChangeDirectory()
	util.LoadEnv()
	database.Init()
	logger.Init()
}

func TestTravel(t *testing.T) {
	u := handler.NewUserHandler()

	from := time.Now().AddDate(0, -1, 0)
	q := model.Query{
		Schema: model.ReadQuery("users"),
		SearchArguments: model.SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
	if err := u.Travel(&from, &q); err != nil {
		t.Error(err.Error())
	}
	if count := database.Count(u.Model.Name()); count == 0 {
		t.Fail()
	}

	DropCollection(u.Model)
}

func TestFetchUsers(t *testing.T) {
	u := handler.NewUserHandler()

	q := model.Query{
		Schema: model.ReadQuery("users"),
		SearchArguments: model.SearchArguments{
			First: 100,
			Query: strconv.Quote("created:2020-01-01..2020-01-01 followers:>=1 repos:>=10 sort:joined"),
			Type:  "USER",
		},
	}

	var users []model.User
	if err := u.FetchUsers(&q, &users); err != nil {
		t.Error(err.Error())
	}
	if len(users) == 0 {
		t.Fail()
	}

	DropCollection(u.Model)
}

func TestStoreUsers(t *testing.T) {
	u := handler.NewUserHandler()

	user := model.User{Login: "memochou1993"}
	users := []model.User{user}
	u.StoreUsers(users)
	if count := database.Count(u.Model.Name()); count == 0 {
		t.Fail()
	}

	DropCollection(u.Model)
}

func TestUpdate(t *testing.T) {
	u := handler.NewUserHandler()

	user := model.User{Login: "memochou1993"}
	users := []model.User{user}
	u.StoreUsers(users)

	if err := u.Update(); err != nil {
		t.Error(err.Error())
	}
	if len(u.GetByLogin(user.Login).Repositories) == 0 {
		t.Fail()
	}

	DropCollection(u.Model)
}

func TestFetchUserRepositories(t *testing.T) {
	u := handler.NewUserHandler()

	q := model.Query{
		Schema: model.ReadQuery("user_repositories"),
		UserArguments: model.UserArguments{
			Login: strconv.Quote("memochou1993"),
		},
		RepositoriesArguments: model.RepositoriesArguments{
			First:             100,
			OrderBy:           "{field:STARGAZERS,direction:DESC}",
			OwnerAffiliations: "OWNER",
		},
	}

	var repos []model.Repository
	if err := u.FetchRepositories(&q, &repos); err != nil {
		t.Error(err.Error())
	}
	if len(repos) == 0 {
		t.Fail()
	}

	DropCollection(u.Model)
}

func TestUpdateRepositories(t *testing.T) {
	u := handler.NewUserHandler()

	user := model.User{Login: "memochou1993"}
	users := []model.User{user}
	u.StoreUsers(users)

	repos := []model.Repository{{Name: "github-rankings"}}
	u.UpdateRepositories(user, repos)
	if len(u.GetByLogin(user.Login).Repositories) == 0 {
		t.Fail()
	}

	DropCollection(u.Model)
}

func TestIndexUsers(t *testing.T) {
	u := handler.NewUserHandler()

	u.CreateIndexes()

	indexes := database.Indexes(u.Model.Name())
	if len(indexes) == 0 {
		t.Fail()
	}

	DropCollection(u.Model)
}

func tearDown() {
	DropDatabase()
}
