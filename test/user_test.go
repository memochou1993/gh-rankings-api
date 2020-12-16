package test

import (
	"github.com/memochou1993/github-rankings/app"
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
	u := app.NewUserModel()

	from := time.Now().AddDate(0, -1, 0)
	q := app.Query{
		Schema: app.ReadQuery("users"),
		SearchArguments: app.SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
	if err := u.Travel(&from, &q); err != nil {
		t.Error(err.Error())
	}
	if count := database.Count(u.Name()); count == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func TestFetchUsers(t *testing.T) {
	u := app.NewUserModel()

	q := app.Query{
		Schema: app.ReadQuery("users"),
		SearchArguments: app.SearchArguments{
			First: 100,
			Query: strconv.Quote("created:2020-01-01..2020-01-01 followers:>=1 repos:>=10 sort:joined"),
			Type:  "USER",
		},
	}

	var users []app.User
	if err := u.FetchUsers(&q, &users); err != nil {
		t.Error(err.Error())
	}
	if len(users) == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func TestStoreUsers(t *testing.T) {
	u := app.NewUserModel()

	user := app.User{Login: "memochou1993"}
	users := []app.User{user}
	u.StoreUsers(users)
	if count := database.Count(u.Name()); count == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func TestUpdate(t *testing.T) {
	u := app.NewUserModel()

	user := app.User{Login: "memochou1993"}
	users := []app.User{user}
	u.StoreUsers(users)

	if err := u.Update(); err != nil {
		t.Error(err.Error())
	}
	if len(u.GetByLogin(user.Login).Repositories) == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func TestFetchUserRepositories(t *testing.T) {
	u := app.NewUserModel()

	q := app.Query{
		Schema: app.ReadQuery("user_repositories"),
		UserArguments: app.UserArguments{
			Login: strconv.Quote("memochou1993"),
		},
		RepositoriesArguments: app.RepositoriesArguments{
			First:             100,
			OrderBy:           "{field:STARGAZERS,direction:DESC}",
			OwnerAffiliations: "OWNER",
		},
	}

	var repos []app.Repository
	if err := u.FetchRepositories(&q, &repos); err != nil {
		t.Error(err.Error())
	}
	if len(repos) == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func TestUpdateRepositories(t *testing.T) {
	u := app.NewUserModel()

	user := app.User{Login: "memochou1993"}
	users := []app.User{user}
	u.StoreUsers(users)

	repos := []app.Repository{{Name: "github-rankings"}}
	u.UpdateRepositories(user, repos)
	if len(u.GetByLogin(user.Login).Repositories) == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func TestIndexUsers(t *testing.T) {
	u := app.NewUserModel()

	u.CreateIndexes()

	indexes := database.Indexes(u.Name())
	if len(indexes) == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func tearDown() {
	DropDatabase()
}
