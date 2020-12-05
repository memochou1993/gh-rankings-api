package model

import (
	"context"
	"github.com/memochou1993/github-rankings/app/database"
	"github.com/memochou1993/github-rankings/app/model"
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

func TestStoreUsers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users := &model.Users{}
	if err := users.Search(ctx); err != nil {
		t.Error(err.Error())
	}
	if err := users.Store(ctx); err != nil {
		t.Error(err.Error())
	}

	count, err := database.Count(ctx, model.CollectionUsers)
	if err != nil {
		t.Error(err.Error())
	}
	if count != 100 {
		t.Fail()
	}
}

func tearDown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := database.GetDatabase().Drop(ctx); err != nil {
		log.Fatal(err.Error())
	}
}
