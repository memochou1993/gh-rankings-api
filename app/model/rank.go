package model

import (
	"time"
)

type Rank struct {
	Rank       int       `bson:"rank"`
	TotalCount int       `bson:"total_count"`
	Tags       []string  `bson:"tags"`
	Batch      int       `bson:"batch"`
	CreatedAt  time.Time `bson:"created_at"`
}
