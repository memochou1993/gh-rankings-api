package model

import (
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/test"
	"github.com/memochou1993/github-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
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
	test.ChangeDirectory()
	util.LoadEnv()
	database.Init()
	logger.Init()
}

func TestTravel(t *testing.T) {
	u := model.NewUserCollection()

	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()
	r := app.Request{
		Schema: app.Read("users"),
		SearchArguments: app.SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
	if err := u.Travel(&from, &to, &r); err != nil {
		t.Error(err.Error())
	}
	if count := u.Count(); count == 0 {
		t.Fail()
	}

	test.DropCollection(u)
}

func TestFetchUsers(t *testing.T) {
	u := model.NewUserCollection()

	r := app.Request{
		Schema: app.Read("users"),
		SearchArguments: app.SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
	r.SearchArguments.Query = r.String("created:2020-01-01..2020-01-01 followers:>=1 repos:>=10")

	var users []interface{}
	if err := u.FetchUsers(&r, &users); err != nil {
		t.Error(err.Error())
	}
	if len(users) == 0 {
		t.Fail()
	}

	test.DropCollection(u)
}

func TestStoreUsers(t *testing.T) {
	u := model.NewUserCollection()

	var users []interface{}
	users = append(users, bson.D{})
	if err := u.StoreUsers(users); err != nil {
		t.Error(err.Error())
	}
	if count := u.Count(); count != 1 {
		t.Fail()
	}

	test.DropCollection(u)
}

func TestIndexUsers(t *testing.T) {
	u := model.NewUserCollection()

	if err := u.Index(); err != nil {
		t.Error(err.Error())
	}

	indexes := database.GetIndexes("users")
	if len(indexes) == 0 {
		t.Fail()
	}

	test.DropCollection(u)
}

func tearDown() {
	test.DropDatabase()
}
