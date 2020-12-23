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
	Rank       int       `bson:"rank"`
	TotalCount int       `bson:"total_count"`
	Tags       []string  `bson:"tags"`
	CreatedAt  time.Time `bson:"created_at"`
}

type RankPipeline struct {
	Pipeline mongo.Pipeline
	Tags     []string
}

func PushRanks(model Interface, batch time.Time, pipeline RankPipeline) {
	ctx := context.Background()
	cursor := database.Aggregate(ctx, model.Name(), pipeline.Pipeline)
	defer database.CloseCursor(ctx, cursor)

	var models []mongo.WriteModel
	for count := 0; cursor.Next(ctx); count++ {
		record := struct {
			ID         string `bson:"_id"`
			TotalCount int    `bson:"total_count"`
		}{}
		if err := cursor.Decode(&record); err != nil {
			log.Fatalln(err.Error())
		}

		rank := Rank{
			Rank:       count + 1,
			TotalCount: record.TotalCount,
			Tags:       pipeline.Tags,
			CreatedAt:  batch,
		}
		filter := bson.D{{"_id", record.ID}}
		update := bson.D{{"$push", bson.D{{"ranks", rank}}}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
		if cursor.RemainingBatchLength() == 0 {
			if _, err := model.Collection().BulkWrite(ctx, models); err != nil {
				log.Fatalln(err.Error())
			}
			models = models[:0]
		}
	}
}

func PullRanks(model Interface, batch time.Time) {
	filter := bson.D{}
	update := bson.D{{"$pull", bson.D{{"ranks", bson.D{{"created_at", bson.D{{"$lt", batch}}}}}}}}
	database.UpdateMany(model.Name(), filter, update)
}
