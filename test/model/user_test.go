package model

import (
	"context"
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
	u := model.UserCollection{}
	u.SetCollectionName("users")

	date := time.Now().AddDate(0, -1, 0)
	r := app.Request{
		Schema: app.Read("users"),
		SearchArguments: app.SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
	if err := u.Travel(&date, &r); err != nil {
		t.Error(err.Error())
	}
	if count := u.Count(); count == 0 {
		t.Fail()
	}

	test.DropCollection(&u)
}

func TestFetchUsers(t *testing.T) {
	u := model.UserCollection{}
	u.SetCollectionName("users")

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

	test.DropCollection(&u)
}

func TestStoreUsers(t *testing.T) {
	u := model.UserCollection{}
	u.SetCollectionName("users")

	var users []interface{}
	users = append(users, bson.D{})
	if err := u.StoreUsers(users); err != nil {
		t.Error(err.Error())
	}
	if count := u.Count(); count != 1 {
		t.Fail()
	}

	test.DropCollection(&u)
}

func TestIndexUsers(t *testing.T) {
	u := model.UserCollection{}
	u.SetCollectionName("users")

	ctx := context.Background()

	if err := u.Index([]string{"login"}); err != nil {
		t.Error(err.Error())
	}

	cursor, err := u.GetCollection().Indexes().List(ctx)
	if err != nil {
		t.Fatal()
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			t.Fatal()
		}
	}()

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		t.Fatal()
	}
	if len(indexes) == 0 {
		t.Fail()
	}

	test.DropCollection(&u)
}

func tearDown() {
	test.DropDatabase()
}
