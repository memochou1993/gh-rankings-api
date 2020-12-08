package model

import (
	"context"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/app/query"
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func TestSearchUsers(t *testing.T) {
	users := model.Users{}
	args := query.SearchArguments{
		First: 1,
		Query: "\"repos:>=5 followers:>=10\"",
		Type:  "USER",
	}
	if err := users.Search(&args); err != nil {
		t.Error(err.Error())
	}
	if len(users.Data.Search.Edges) != 1 {
		t.Fail()
	}
}

func TestStoreUsers(t *testing.T) {
	users := model.Users{}
	args := query.SearchArguments{
		First: 1,
		Query: "\"repos:>=5 followers:>=10\"",
		Type:  "USER",
	}
	if err := users.Search(&args); err != nil {
		t.Error(err.Error())
	}
	if err := users.Store(); err != nil {
		t.Error(err.Error())
	}

	count, err := database.Count(context.Background(), model.CollectionUsers)
	if err != nil {
		t.Error(err.Error())
	}
	if count != 1 {
		t.Fail()
	}

	dropCollection()
}

func TestIndexUsers(t *testing.T) {
	users := model.Users{}
	args := query.SearchArguments{
		First: 1,
		Query: "\"repos:>=5 followers:>=10\"",
		Type:  "USER",
	}
	if err := users.Search(&args); err != nil {
		t.Error(err.Error())
	}
	if err := users.Store(); err != nil {
		t.Error(err.Error())
	}

	cursor, err := database.GetCollection(model.CollectionUsers).Indexes().List(context.Background())
	if err != nil {
		t.Error(err.Error())
		return
	}

	var indexes []bson.M
	if err := cursor.All(context.Background(), &indexes); err != nil {
		log.Fatal(err)
	}
	if len(indexes) == 0 {
		t.Fail()
	}

	dropCollection()
}

func dropCollection() {
	if err := database.GetCollection(model.CollectionUsers).Drop(context.Background()); err != nil {
		log.Fatal(err.Error())
	}
}

func tearDown() {
	if err := database.GetDatabase().Drop(context.Background()); err != nil {
		log.Fatal(err.Error())
	}
}
