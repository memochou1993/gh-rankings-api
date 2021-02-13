package model

import (
	"context"
	"fmt"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/pipeline"
	"github.com/memochou1993/gh-rankings/database"
	"github.com/memochou1993/gh-rankings/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type Rank struct {
	Name      string    `json:"name" bson:"name"`
	ImageUrl  string    `json:"imageUrl" bson:"image_url"`
	Rank      int       `json:"rank" bson:"rank"`
	RankCount int       `json:"rankCount" bson:"rank_count"`
	ItemCount int       `json:"itemCount" bson:"item_count"`
	Type      string    `json:"type" bson:"type"`
	Field     string    `json:"field" bson:"field"`
	Language  string    `json:"language" bson:"language"`
	Location  string    `json:"location" bson:"location"`
	CreatedAt time.Time `json:"createdAt" bson:"created_at"`
}

type RankModel struct {
	*Model
}

func (r *RankModel) CreateIndexes() {
	indexes := []string{"name", "type", "field", "language", "location", "created_at"}
	database.CreateIndexes(r.Name(), indexes)
	logger.Success(fmt.Sprintf("Created %d indexes on %s collection!", len(indexes), r.Name()))
}

func (r *RankModel) List(req *request.Request, createdAt time.Time) []Rank {
	ctx := context.Background()

	// TODO: by name
	// cond := mongo.Pipeline{{
	// 	{"created_at", createdAt},
	// 	{"name", req.Name},
	// }}
	// if req.Type != "" {
	// 	cond = append(cond, bson.D{{"type", req.Type}})
	// }
	// if req.Field != "" {
	// 	cond = append(cond, bson.D{{"type", req.Field}})
	// }
	// if req.Language != "" {
	// 	cond = append(cond, bson.D{{"type", req.Language}})
	// }
	// if req.Location != "" {
	// 	cond = append(cond, bson.D{{"type", req.Location}})
	// }

	p := pipeline.List(req, createdAt)
	cursor := database.Aggregate(ctx, r.Model.Name(), p)
	ranks := make([]Rank, req.Limit)
	if err := cursor.All(ctx, &ranks); err != nil {
		log.Fatal(err.Error())
	}
	return ranks
}

func (r *RankModel) Store(model Interface, p pipeline.Pipeline, createdAt time.Time) {
	ctx := context.Background()
	cursor := database.Aggregate(ctx, model.Name(), *p.Pipeline)
	defer database.CloseCursor(ctx, cursor)

	count := r.Count(model, p)

	var models []mongo.WriteModel
	for i := 0; cursor.Next(ctx); i++ {
		rec := struct {
			ID         string `bson:"_id"`
			ImageUrl   string `bson:"image_url"`
			TotalCount int    `bson:"total_count"`
		}{}
		if err := cursor.Decode(&rec); err != nil {
			log.Fatal(err.Error())
		}

		rank := Rank{
			Name:      rec.ID,
			ImageUrl:  rec.ImageUrl,
			Rank:      i + 1,
			RankCount: count,
			ItemCount: rec.TotalCount,
			Type:      p.Type,
			Field:     p.Field,
			Language:  p.Language,
			Location:  p.Location,
			CreatedAt: createdAt,
		}
		models = append(models, mongo.NewInsertOneModel().SetDocument(rank))
		if cursor.RemainingBatchLength() == 0 {
			database.BulkWrite(r.Name(), models)
			models = models[:0]
		}
	}
}

func (r *RankModel) Delete(createdAt time.Time, rankType string) {
	filter := bson.D{
		{"$and", []bson.D{{
			{"type", rankType},
			{"created_at", bson.D{
				{"$lt", createdAt},
			}},
		}}},
	}
	database.DeleteMany(r.Name(), filter)
}

func (r *RankModel) Count(model Interface, p pipeline.Pipeline) int {
	ctx := context.Background()
	rec := struct {
		Count int `bson:"count"`
	}{}
	cursor := database.Aggregate(ctx, model.Name(), pipeline.RankCount(*p.Pipeline))
	defer database.CloseCursor(ctx, cursor)
	for cursor.Next(ctx) {
		if err := cursor.Decode(&rec); err != nil {
			log.Fatal(err.Error())
		}
	}
	return rec.Count
}

func NewRankModel(name string) *RankModel {
	return &RankModel{
		&Model{
			name: fmt.Sprintf("%s_ranks", name),
		},
	}
}
