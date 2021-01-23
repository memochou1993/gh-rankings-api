package model

import (
	"context"
	"github.com/memochou1993/github-rankings/app/handler/request"
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strings"
	"time"
)

type Rank struct {
	Title      string    `json:"title" bson:"title"`
	ImageUrl   string    `json:"imageUrl" bson:"image_url"`
	Rank       int       `json:"rank" bson:"rank"`
	Last       int       `json:"last" bson:"last"`
	TotalCount int       `json:"totalCount" bson:"total_count"`
	Tags       []string  `json:"tags" bson:"tags"`
	CreatedAt  time.Time `json:"createdAt" bson:"created_at"`
}

type RankRecord struct {
	ID         string `bson:"_id"`
	ImageUrl   string `bson:"image_url"`
	TotalCount int    `bson:"total_count"`
}

type RankModel struct {
	*Model
}

func (r *RankModel) CreateIndexes() {
	database.CreateIndexes(r.Name(), []string{
		"title",
		"tags",
	})
}

func (r *RankModel) List(req *request.Request, ranks *[]Rank) {
	ctx := context.Background()
	cond := mongo.Pipeline{bson.D{{"created_at", req.Timestamp}}}
	if req.Title != "" {
		cond = append(cond, bson.D{{"title", req.Title}})
	}
	if strings.Join(req.Tags, "") != "" {
		cond = append(cond, bson.D{{"tags", req.Tags}})
	}
	pipeline := mongo.Pipeline{
		bson.D{
			{"$match", bson.D{
				{"$and", cond},
			}},
		},
		bson.D{
			{"$skip", (req.Page - 1) * req.Limit},
		},
		bson.D{
			{"$limit", req.Limit},
		},
	}
	cursor := database.Aggregate(ctx, NewRankModel().Name(), pipeline)
	if err := cursor.All(context.Background(), ranks); err != nil {
		log.Fatal(err.Error())
	}
}

func (r *RankModel) Store(model Interface, p Pipeline, createdAt time.Time) int {
	ctx := context.Background()
	cursor := database.Aggregate(ctx, model.Name(), *p.Pipeline)
	defer database.CloseCursor(ctx, cursor)

	count := p.Count(model)

	var models []mongo.WriteModel
	for i := 0; cursor.Next(ctx); i++ {
		rec := RankRecord{}
		if err := cursor.Decode(&rec); err != nil {
			log.Fatal(err.Error())
		}

		doc := Rank{
			Title:      rec.ID,
			ImageUrl:   rec.ImageUrl,
			Rank:       i + 1,
			Last:       count,
			TotalCount: rec.TotalCount,
			Tags:       p.Tags,
			CreatedAt:  createdAt,
		}
		models = append(models, mongo.NewInsertOneModel().SetDocument(doc))
		if cursor.RemainingBatchLength() == 0 {
			database.BulkWrite(r.Name(), models)
			models = models[:0]
		}
	}
	return count
}

func (r *RankModel) Delete(createdAt time.Time, tags ...string) {
	filter := bson.D{
		{"$and", []bson.D{{
			{"tags", bson.D{
				{"$in", tags},
			}},
			{"created_at", bson.D{
				{"$lt", createdAt},
			}},
		}}},
	}
	database.DeleteMany(r.Name(), filter)
}

type Pipeline struct {
	Pipeline *mongo.Pipeline
	Tags     []string
}

func (p *Pipeline) Count(model Interface) int {
	ctx := context.Background()
	rec := struct {
		Count int `bson:"count"`
	}{}
	cursor := database.Aggregate(ctx, model.Name(), append(*p.Pipeline, bson.D{{"$count", "count"}}))
	defer database.CloseCursor(ctx, cursor)
	for cursor.Next(ctx) {
		if err := cursor.Decode(&rec); err != nil {
			log.Fatal(err.Error())
		}
	}
	return rec.Count
}

func NewRankModel() *RankModel {
	return &RankModel{
		&Model{
			name: "ranks",
		},
	}
}
