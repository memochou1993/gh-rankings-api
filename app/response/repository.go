package response

import "github.com/memochou1993/gh-rankings/app/model"

type Repository struct {
	Message string `json:"message"`
	Data    struct {
		Search struct {
			Edges []struct {
				Cursor string           `json:"cursor"`
				Node   model.Repository `json:"node"`
			} `json:"edges"`
			PageInfo `json:"pageInfo"`
		} `json:"search"`
		RateLimit `json:"rateLimit"`
	} `json:"data"`
	Errors []Error `json:"errors"`
}
