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
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}

type Pipeline struct {
	Pipeline *mongo.Pipeline
	Tags     []string
}

func CountRanks(model Interface, p mongo.Pipeline) int {
	ctx := context.Background()
	r := struct {
		Count int `bson:"count"`
	}{}
	cursor := database.Aggregate(ctx, model.Name(), append(p, bson.D{{"$count", "count"}}))
	defer database.CloseCursor(ctx, cursor)
	for cursor.Next(ctx) {
		if err := cursor.Decode(&r); err != nil {
			log.Fatalln(err.Error())
		}
	}
	return r.Count
}

func PushRanks(model Interface, timestamp time.Time, p Pipeline) {
	ctx := context.Background()
	cursor := database.Aggregate(ctx, model.Name(), *p.Pipeline)
	defer database.CloseCursor(ctx, cursor)

	last := CountRanks(model, *p.Pipeline)

	var models []mongo.WriteModel
	for i := 0; cursor.Next(ctx); i++ {
		r := struct {
			ID         string `bson:"_id"`
			TotalCount int    `bson:"total_count"`
		}{}
		if err := cursor.Decode(&r); err != nil {
			log.Fatalln(err.Error())
		}

		rank := Rank{
			Rank:       i + 1,
			Last:       last,
			TotalCount: r.TotalCount,
			Tags:       p.Tags,
			UpdatedAt:  timestamp,
		}
		filter := bson.D{{"_id", r.ID}}
		update := bson.D{{"$push", bson.D{{"ranks", rank}}}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
		if cursor.RemainingBatchLength() == 0 {
			database.BulkWrite(model.Name(), models)
			models = models[:0]
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func PullRanks(model Interface, timestamp time.Time) {
	filter := bson.D{}
	update := bson.D{{"$pull", bson.D{{"ranks", bson.D{{"updated_at", bson.D{{"$lt", timestamp}}}}}}}}
	database.UpdateMany(model.Name(), filter, update)
}
