package response

import (
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/query"
	"time"
)

type User struct {
	Data struct {
		Search struct {
			Edges []struct {
				Cursor string     `json:"cursor"`
				Node   model.User `json:"node"`
			} `json:"edges"`
			PageInfo `json:"pageInfo"`
		} `json:"search"`
		User struct {
			ImageUrl  string      `json:"imageUrl"`
			CreatedAt time.Time   `json:"createdAt"`
			Followers query.Items `json:"followers"`
			Gists     struct {
				Edges []struct {
					Cursor string     `json:"cursor"`
					Node   query.Gist `json:"node"`
				} `json:"edges"`
				PageInfo `json:"pageInfo"`
			} `json:"gists"`
			Location     string `json:"location"`
			Login        string `json:"login"`
			Name         string `json:"name"`
			Repositories struct {
				Edges []struct {
					Cursor string           `json:"cursor"`
					Node   model.Repository `json:"node"`
				} `json:"edges"`
				PageInfo `json:"pageInfo"`
			} `json:"repositories"`
		} `json:"owner"`
		RateLimit `json:"rateLimit"`
	} `json:"data"`
	Errors  []Error `json:"errors"`
	Message string  `json:"message"`
}
