package model

import (
	"context"
	"github.com/memochou1993/github-rankings/app/database"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/app/query"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func TestSearchUsers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users := &model.Users{}
	args := &query.SearchArguments{
		First: 1,
		Query: "\"repos:>=5 followers:>=10\"",
		Type:  "USER",
	}
	if err := users.Search(ctx, args); err != nil {
		t.Error(err.Error())
	}
	if len(users.Data.Search.Edges) != 1 {
		t.Fail()
	}
}

func TestStoreUsers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users := &model.Users{}
	args := &query.SearchArguments{
		First: 1,
		Query: "\"repos:>=5 followers:>=10\"",
		Type:  "USER",
	}
	if err := users.Search(ctx, args); err != nil {
		t.Error(err.Error())
	}
	if err := users.Store(ctx); err != nil {
		t.Error(err.Error())
	}

	count, err := database.Count(ctx, model.CollectionUsers)
	if err != nil {
		t.Error(err.Error())
	}
	if count != 1 {
		t.Fail()
	}

	dropCollection(ctx)
}

func TestIndexUsers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users := &model.Users{}
	args := &query.SearchArguments{
		First: 1,
		Query: "\"repos:>=5 followers:>=10\"",
		Type:  "USER",
	}
	if err := users.Search(ctx, args); err != nil {
		t.Error(err.Error())
	}
	if err := users.Store(ctx); err != nil {
		t.Error(err.Error())
	}

	cursor, err := database.GetCollection(model.CollectionUsers).Indexes().List(ctx)
	if err != nil {
		t.Error(err.Error())
		return
	}

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		log.Fatal(err)
	}
	if len(indexes) == 0 {
		t.Fail()
	}

	dropCollection(ctx)
}

func dropCollection(ctx context.Context) {
	if err := database.GetCollection(model.CollectionUsers).Drop(ctx); err != nil {
		log.Fatal(err.Error())
	}
}

func tearDown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := database.GetDatabase().Drop(ctx); err != nil {
		log.Fatal(err.Error())
	}
}
