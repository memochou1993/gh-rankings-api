package test

import (
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"os"
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
	u := app.NewUserCollection()

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
	if count := database.Count(u.GetName()); count == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func TestFetchUsers(t *testing.T) {
	u := app.NewUserCollection()

	q := app.Query{
		Schema: app.ReadQuery("users"),
		SearchArguments: app.SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
	q.SearchArguments.Query = q.String("created:2020-01-01..2020-01-01 followers:>=1 repos:>=10 sort:joined")

	var users []interface{}
	if err := u.FetchUsers(&q, &users); err != nil {
		t.Error(err.Error())
	}
	if len(users) == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func TestStoreUsers(t *testing.T) {
	u := app.NewUserCollection()

	users := []interface{}{app.User{}}
	u.StoreUsers(users)
	if count := database.Count(u.GetName()); count == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func TestUpdate(t *testing.T) {
	u := app.NewUserCollection()

	user := app.User{
		Login: "memochou1993",
	}

	users := []interface{}{user}
	u.StoreUsers(users)

	if err := u.Update(); err != nil {
		t.Error(err.Error())
	}
	if len(u.GetLast().Repositories) == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func TestFetchUserRepositories(t *testing.T) {
	u := app.NewUserCollection()

	q := app.Query{
		Schema: app.ReadQuery("user_repositories"),
		RepositoriesArguments: app.RepositoriesArguments{
			First:             100,
			OrderBy:           "{field:STARGAZERS,direction:DESC}",
			OwnerAffiliations: "OWNER",
		},
	}
	q.UserArguments.Login = q.String("memochou1993")

	var repos []interface{}
	if err := u.FetchRepositories(&q, &repos); err != nil {
		t.Error(err.Error())
	}
	if len(repos) == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func TestUpdateRepositories(t *testing.T) {
	u := app.NewUserCollection()

	user := app.User{
		Login: "memochou1993",
	}

	users := []interface{}{user}
	u.StoreUsers(users)

	repos := []interface{}{app.Repository{}}
	u.UpdateRepositories(user, repos)
	if len(u.GetLast().Repositories) == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func TestIndexUsers(t *testing.T) {
	u := app.NewUserCollection()

	u.CreateIndexes()

	indexes := database.GetIndexes(u.GetName())
	if len(indexes) == 0 {
		t.Fail()
	}

	DropCollection(u)
}

func tearDown() {
	DropDatabase()
}
