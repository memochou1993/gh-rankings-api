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
	userCollection := model.UserCollection{}
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
	args := query.SearchArguments{
		First: 1,
		Query: "\"repos:>=5 followers:>=10\"",
		Type:  "USER",
	}
	if err := userCollection.Search(&args); err != nil {
		t.Error(err.Error())
	}

	var users []interface{}
	for _, edge := range userCollection.SearchResult.Data.Search.Edges {
		users = append(users, bson.D{
			{"login", edge.Node.Login},
			{"name", edge.Node.Name},
		})
	}
	if err := userCollection.Store(users); err != nil {
		t.Error(err.Error())
	}

	count, err := userCollection.Count()
	if err != nil {
		t.Error(err.Error())
	}
	if count != 1 {
		t.Fail()
	}

	dropCollection()
}

func TestIndexUsers(t *testing.T) {
	ctx := context.Background()

	users := model.UserCollection{}
	if err := users.Index(); err != nil {
		t.Error(err.Error())
	}

	cursor, err := database.GetCollection(model.CollectionUsers).Indexes().List(ctx)
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			t.Fatal()
		}
	}()

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		t.Error(err.Error())
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
