package model

import (
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Rank struct {
	Rank       int       `bson:"rank"`
	TotalCount int       `bson:"total_count"`
	Tags       []string  `bson:"tags"`
	Batch      int       `bson:"batch"`
	CreatedAt  time.Time `bson:"created_at"`
}

type RankPipeline struct {
	Pipeline mongo.Pipeline
	Tags     []string
}
