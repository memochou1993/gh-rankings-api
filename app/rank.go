package app

import (
	"time"
)

type Rank struct {
	Rank       int       `bson:"rank"`
	TotalCount int       `bson:"total_count"`
	CreatedAt  time.Time `bson:"created_at"`
}
