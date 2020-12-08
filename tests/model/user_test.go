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
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	//
}

func TestSearchUsers(t *testing.T) {
	userCollection := model.UserCollection{}
	if err := userCollection.Init(); err != nil {
		t.Error(err.Error())
	}

	args := query.SearchArguments{
		First: 1,
		Query: "\"repos:>=5 followers:>=10\"",
		Type:  "USER",
	}
	if err := userCollection.Search(&args); err != nil {
		t.Error(err.Error())
	}
	if len(userCollection.SearchResult.Data.Search.Edges) != 1 {
		t.Fail()
	}
}

func TestStoreUsers(t *testing.T) {
	userCollection := model.UserCollection{}
	if err := userCollection.Init(); err != nil {
		t.Error(err.Error())
	}

	args := query.SearchArguments{
		First: 1,
		Query: "\"repos:>=5 followers:>=10\"",
		Type:  "USER",
	}
	if err := userCollection.Search(&args); err != nil {
		t.Error(err.Error())
	}
	if err := userCollection.StoreSearchResult(); err != nil {
		t.Error(err.Error())
	}

	count, err := userCollection.Count()
	if err != nil {
		t.Error(err.Error())
	}
	if count != 1 {
		t.Fail()
	}

	dropCollection(&userCollection)
}

func TestIndexUsers(t *testing.T) {
	userCollection := model.UserCollection{}
	if err := userCollection.Init(); err != nil {
		t.Error(err.Error())
	}

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

func dropCollection(collection model.CollectionInterface) {
	if err := collection.GetCollection().Drop(context.Background()); err != nil {
		log.Fatal(err.Error())
	}
}

func tearDown() {
	if err := database.GetDatabase().Drop(context.Background()); err != nil {
		log.Fatal(err.Error())
	}
}
