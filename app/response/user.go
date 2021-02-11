package response

import (
	"github.com/memochou1993/gh-rankings/app/model"
	"time"
)

type User struct {
	Message string `json:"message"`
	Data    struct {
		Search struct {
			Edges []struct {
				Cursor string     `json:"cursor"`
				Node   model.User `json:"node"`
			} `json:"edges"`
			PageInfo `json:"pageInfo"`
		} `json:"search"`
		User struct {
			ImageUrl  string       `json:"imageUrl"`
			CreatedAt *time.Time   `json:"createdAt"`
			Followers *model.Items `json:"followers"`
			Gists     struct {
				Edges []struct {
					Cursor string     `json:"cursor"`
					Node   model.Gist `json:"node"`
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
	Errors []Error `json:"errors"`
}
