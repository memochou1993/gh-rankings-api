package model

import (
	"context"
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strings"
	"time"
)

type OwnerRank struct {
	Login string `json:"login" bson:"login"`
	Rank  *Rank  `json:"rank,omitempty" bson:"rank,omitempty"`
}

type OwnerRankModel struct {
	*Model
}

type OwnerRankArguments struct {
	Login     string
	Tags      []string
	CreatedAt time.Time
	Page      int
}

func (o *OwnerRankModel) CreateIndexes() {
	database.CreateIndexes(o.Name(), []string{
		"login",
		"rank.tags",
	})
}

func (o *OwnerRankModel) List(args OwnerRankArguments) *mongo.Cursor {
	ctx := context.Background()
	limit := 10
	match := mongo.Pipeline{bson.D{{"rank.created_at", args.CreatedAt}}}
	if args.Login != "" {
		match = append(match, bson.D{{"login", args.Login}})
	}
	if strings.Join(args.Tags, "") != "" {
		match = append(match, bson.D{{"rank.tags", args.Tags}})
	}
	pipeline := mongo.Pipeline{
		bson.D{
			{"$match", bson.D{
				{"$and", match},
			}},
		},
		bson.D{
			{"$skip", (args.Page - 1) * limit},
		},
		bson.D{
			{"$limit", limit},
		},
	}
	return database.Aggregate(ctx, NewOwnerRankModel().Name(), pipeline)
}

func (o *OwnerRankModel) Store(createdAt time.Time, p Pipeline) {
	ctx := context.Background()
	model := NewOwnerModel()
	cursor := database.Aggregate(ctx, model.Name(), *p.Pipeline)
	defer database.CloseCursor(ctx, cursor)

	last := p.Count(model)

	var models []mongo.WriteModel
	for i := 0; cursor.Next(ctx); i++ {
		rec := struct {
			ID         string `bson:"_id"`
			TotalCount int    `bson:"total_count"`
		}{}
		if err := cursor.Decode(&rec); err != nil {
			log.Fatalln(err.Error())
		}

		doc := OwnerRank{
			Login: rec.ID,
			Rank: &Rank{
				Rank:       i + 1,
				Last:       last,
				TotalCount: rec.TotalCount,
				Tags:       p.Tags,
				CreatedAt:  createdAt,
			},
		}
		models = append(models, mongo.NewInsertOneModel().SetDocument(doc))
		if cursor.RemainingBatchLength() == 0 {
			database.BulkWrite(o.Name(), models)
			models = models[:0]
		}
	}
}

func (o *OwnerRankModel) Delete(createdAt time.Time) {
	database.DeleteMany(o.Name(), bson.D{{"rank.created_at", bson.D{{"$lt", createdAt}}}})
}

func NewOwnerRankModel() *OwnerRankModel {
	return &OwnerRankModel{
		&Model{
			name: "owner_ranks",
		},
	}
}
