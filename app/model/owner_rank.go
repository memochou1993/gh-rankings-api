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

type OwnerRank struct {
	AvatarURL string `json:"avatarUrl" bson:"avatar_url"`
	Login     string `json:"login" bson:"login"`
	Rank      *Rank  `json:"rank" bson:"rank"`
}

type OwnerRankRecord struct {
	ID         string `bson:"_id"`
	AvatarURL  string `json:"avatarUrl" bson:"avatar_url"`
	TotalCount int    `bson:"total_count"`
}

type OwnerRankModel struct {
	*Model
}

func (o *OwnerRankModel) CreateIndexes() {
	database.CreateIndexes(o.Name(), []string{
		"login",
		"rank.tags",
	})
}

func (o *OwnerRankModel) List(req *request.OwnerRequest) *mongo.Cursor {
	ctx := context.Background()
	match := mongo.Pipeline{bson.D{{"rank.created_at", req.CreatedAt}}}
	if req.Login != "" {
		match = append(match, bson.D{{"login", req.Login}})
	}
	if strings.Join(req.Tags, "") != "" {
		match = append(match, bson.D{{"rank.tags", req.Tags}})
	}
	pipeline := mongo.Pipeline{
		bson.D{
			{"$match", bson.D{
				{"$and", match},
			}},
		},
		bson.D{
			{"$skip", (req.Page - 1) * req.Limit},
		},
		bson.D{
			{"$limit", req.Limit},
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
		rec := OwnerRankRecord{}
		if err := cursor.Decode(&rec); err != nil {
			log.Fatalln(err.Error())
		}

		doc := OwnerRank{
			Login:     rec.ID,
			AvatarURL: rec.AvatarURL,
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
