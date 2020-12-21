package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
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

func TotalCountPipeline(ownerType string, object string) mongo.Pipeline {
	return mongo.Pipeline{
		bson.D{
			{"$match", bson.D{
				{"type", ownerType},
			}},
		},
		bson.D{
			{"$project", bson.D{
				{"_id", "$_id"},
				{"total_count", bson.D{
					{"$sum", fmt.Sprintf("$%s.total_count", object)},
				}},
			}},
		},
		bson.D{
			{"$sort", bson.D{
				{"total_count", -1},
			}},
		},
	}
}
