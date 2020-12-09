package model

import (
	"context"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/app/query"
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
	//
}

func TestTravel(t *testing.T) {
	userCollection := model.UserCollection{}
	userCollection.SetCollectionName("users")

	date := time.Now().AddDate(0, -1, 0)
	request := query.Request{
		Schema: query.Read("users"),
		SearchArguments: query.SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
	if err := userCollection.Travel(&date, &request); err != nil {
		t.Error(err.Error())
	}
	if count := userCollection.Count(); count == 0 {
		t.Fail()
	}

	dropCollection(&userCollection)
}

func TestFetchUsers(t *testing.T) {
	userCollection := model.UserCollection{}
	userCollection.SetCollectionName("users")

	request := query.Request{
		Schema: query.Read("users"),
		SearchArguments: query.SearchArguments{
			First: 100,
			Query: "\"created:2020-01-01..2020-01-01 followers:>=1 repos:>=5\"",
			Type:  "USER",
		},
	}
	if err := userCollection.FetchUsers(&request); err != nil {
		t.Error(err.Error())
	}
	if count := userCollection.Count(); count == 0 {
		t.Fail()
	}

	dropCollection(&userCollection)
}

func TestFetch(t *testing.T) {
	userCollection := model.UserCollection{}
	userCollection.SetCollectionName("users")

	request := query.Request{
		Schema: query.Read("users"),
		SearchArguments: query.SearchArguments{
			First: 1,
			Query: "\"repos:>=1 followers:>=1\"",
			Type:  "USER",
		},
	}

	if err := userCollection.Fetch(request.Join()); err != nil {
		t.Error(err.Error())
	}
	if len(userCollection.Response.Data.Search.Edges) != 1 {
		t.Fail()
	}

	dropCollection(&userCollection)
}

func TestStoreUsers(t *testing.T) {
	userCollection := model.UserCollection{}
	userCollection.SetCollectionName("users")

	request := query.Request{
		Schema: query.Read("users"),
		SearchArguments: query.SearchArguments{
			First: 1,
			Query: "\"repos:>=1 followers:>=1\"",
			Type:  "USER",
		},
	}

	if err := userCollection.Fetch(request.Join()); err != nil {
		t.Error(err.Error())
	}
	if err := userCollection.StoreUsers(); err != nil {
		t.Error(err.Error())
	}
	if count := userCollection.Count(); count != 1 {
		t.Fail()
	}

	dropCollection(&userCollection)
}

func TestIndexUsers(t *testing.T) {
	userCollection := model.UserCollection{}
	userCollection.SetCollectionName("users")

	ctx := context.Background()

	if err := userCollection.Index([]string{"login"}); err != nil {
		t.Error(err.Error())
	}

	cursor, err := userCollection.GetCollection().Indexes().List(ctx)
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

	dropCollection(&userCollection)
}

func tearDown() {
	dropDatabase()
}
