package model

import (
	"context"
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type Rank struct {
	Rank       int       `json:"rank" bson:"rank"`
	Last       int       `json:"last" bson:"last"`
	TotalCount int       `json:"total_count" bson:"total_count"`
	Tags       []string  `json:"tags" bson:"tags"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
}

type RankModel struct{}

type Pipeline struct {
	Pipeline *mongo.Pipeline
	Tags     []string
}

func (p *Pipeline) Count(model Interface) int {
	ctx := context.Background()
	r := struct {
		Count int `bson:"count"`
	}{}
	cursor := database.Aggregate(ctx, model.Name(), append(*p.Pipeline, bson.D{{"$count", "count"}}))
	defer database.CloseCursor(ctx, cursor)
	for cursor.Next(ctx) {
		if err := cursor.Decode(&r); err != nil {
			log.Fatalln(err.Error())
		}
	}
	return r.Count
}
