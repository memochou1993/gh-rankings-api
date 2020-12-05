package tests

import (
	"context"
	"github.com/memochou1993/github-rankings/app/model"
	"testing"
	"time"
)

func TestSearchUsers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users := &model.Users{}
	if err := users.SearchUsers(ctx);err != nil {
		t.Error(err.Error())
		return
	}
	if len(users.Data.Search.Edges) != 100 {
		t.Fail()
	}
}
