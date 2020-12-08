package model

import (
	"context"
	"fmt"
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/app/query"
	"github.com/memochou1993/github-rankings/util"
	"log"
	"time"
)

type UserCollection struct {
	Collection
	SearchResult struct {
		Data struct {
			Search struct {
				UserCount int `json:"userCount"`
				Edges     []struct {
					Cursor string `json:"cursor"`
					Node   User   `json:"node"`
				} `json:"edges"`
				PageInfo query.PageInfo `json:"pageInfo"`
			} `json:"search"`
			RateLimit query.RateLimit `json:"rateLimit"`
		} `json:"data"`
		Errors []query.Error `json:"errors"`
	}
}

type User struct {
	AvatarURL string    `json:"avatarUrl"`
	CreatedAt time.Time `json:"createdAt"`
	Email     string    `json:"email"`
	Followers struct {
		TotalCount int `json:"totalCount"`
	} `json:"followers"`
	Location string `json:"location"`
	Login    string `json:"login"`
	Name     string `json:"name"`
}

func (u *UserCollection) Init() error {
	u.SetCollectionName("users")

	if u.Count() > 0 {
		return nil
	}
	if err := u.Index([]string{"login"}); err != nil {
		return err
	}

	return nil
}

func (u *UserCollection) Collect() error {
	if u.Count() > 0 {
		return nil
	}

	layout := "2006-01-02"
	date := time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	for ; date.Before(time.Now()); date.AddDate(0, 0, 7) {
		created := fmt.Sprintf("%s..%s", date.Format(layout), date.AddDate(0, 0, 6).Format(layout))
		followers := ">=10"
		repos := ">=5"
		args := query.SearchArguments{
			First: 100,
			Query: fmt.Sprintf("\"created:%s followers:%s repos:%s\"", created, followers, repos),
			Type:  "USER",
		}
		for {
			u.SearchResult.Data.RateLimit.Check()
			if u.SearchResult.Data.Search.PageInfo.EndCursor != "" {
				args.After = fmt.Sprintf("\"%s\"", u.SearchResult.Data.Search.PageInfo.EndCursor)
			}
			util.LogStruct("Search Arguments", args)
			if err := u.Search(&args); err != nil {
				return err
			}
			if len(u.SearchResult.Errors) > 0 {
				util.LogStruct("Errors", u.SearchResult.Errors)
			}
			util.LogStruct("Rate Limit", u.SearchResult.Data.RateLimit)
			if len(u.SearchResult.Data.Search.Edges) == 0 {
				break
			}
			if err := u.StoreSearchResult(); err != nil {
				return err
			}
			log.Println(fmt.Sprintf("Discovered %d users", len(u.SearchResult.Data.Search.Edges)))
			if !u.SearchResult.Data.Search.PageInfo.HasNextPage {
				u.SearchResult.Data.Search.PageInfo.EndCursor = ""
				break
			}
		}
		date = date.AddDate(0, 0, 7)
	}

	return nil
}

// TODO
// func (u *UserCollection) Update() error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
//
// 	cursor , err := u.GetCollection().Find(ctx, bson.M{})
// 	if err != nil {
// 		return err
// 	}
// 	defer func() {
// 		if err := cursor.Close(ctx); err != nil {
// 			log.Fatalln(err.Error())
// 		}
// 	}()
//
// 	for cursor.Next(ctx) {
// 		user := User{}
// 		if err := cursor.Decode(&user); err != nil {
// 			log.Fatalln(err.Error())
// 		}
// 		fmt.Println(user.Login)
// 	}
//
// 	return nil
// }

func (u *UserCollection) StoreSearchResult() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var documents []interface{}
	for _, edge := range u.SearchResult.Data.Search.Edges {
		documents = append(documents, edge.Node)
	}

	_, err := u.GetCollection().InsertMany(ctx, documents)

	return err
}

func (u *UserCollection) Search(args *query.SearchArguments) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return app.Fetch(ctx, []byte(args.Read("users")), &u.SearchResult)
}
