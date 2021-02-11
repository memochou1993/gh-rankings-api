package response

import (
	"github.com/memochou1993/gh-rankings/app/model"
	"time"
)

type Organization struct {
	Message string `json:"message"`
	Data    struct {
		Search struct {
			Edges []struct {
				Cursor string             `json:"cursor"`
				Node   model.Organization `json:"node"`
			} `json:"edges"`
			PageInfo `json:"pageInfo"`
		} `json:"search"`
		Organization struct {
			AvatarURL    string     `json:"avatarUrl"`
			CreatedAt    *time.Time `json:"createdAt"`
			Location     string     `json:"location"`
			Login        string     `json:"login"`
			Name         string     `json:"name"`
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
