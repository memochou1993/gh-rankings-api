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

type RepositoryRank struct {
	NameWithOwner     string `json:"nameWithOwner" bson:"nameWithOwner"`
	OpenGraphImageUrl string `json:"openGraphImageUrl" bson:"open_graph_image_url"`
	Rank              *Rank  `json:"rank,omitempty" bson:"rank,omitempty"`
}

type RepositoryRankRecord struct {
	ID                string `bson:"_id"`
	OpenGraphImageUrl string `bson:"open_graph_image_url"`
	TotalCount        int    `bson:"total_count"`
}

type RepositoryRankModel struct {
	*Model
}

func (r *RepositoryRankModel) CreateIndexes() {
	database.CreateIndexes(r.Name(), []string{
		"nameWithOwner",
		"rank.tags",
	})
}

func (r *RepositoryRankModel) List(req *request.RepositoryRequest) *mongo.Cursor {
	ctx := context.Background()
	match := mongo.Pipeline{bson.D{{"rank.created_at", req.CreatedAt}}}
	if req.NameWithOwner != "" {
		match = append(match, bson.D{{"nameWithOwner", req.NameWithOwner}})
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
	return database.Aggregate(ctx, NewRepositoryRankModel().Name(), pipeline)
}

func (r *RepositoryRankModel) Store(createdAt time.Time, p Pipeline) {
	ctx := context.Background()
	model := NewRepositoryModel()
	cursor := database.Aggregate(ctx, model.Name(), *p.Pipeline)
	defer database.CloseCursor(ctx, cursor)

	last := p.Count(model)

	var models []mongo.WriteModel
	for i := 0; cursor.Next(ctx); i++ {
		rec := RepositoryRankRecord{}
		if err := cursor.Decode(&rec); err != nil {
			log.Fatalln(err.Error())
		}

		doc := RepositoryRank{
			NameWithOwner:     rec.ID,
			OpenGraphImageUrl: rec.OpenGraphImageUrl,
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
			database.BulkWrite(r.Name(), models)
			models = models[:0]
		}
	}
}

func (r *RepositoryRankModel) Delete(createdAt time.Time) {
	database.DeleteMany(r.Name(), bson.D{{"rank.created_at", bson.D{{"$lt", createdAt}}}})
}

func NewRepositoryRankModel() *RepositoryRankModel {
	return &RepositoryRankModel{
		&Model{
			name: "repository_ranks",
		},
	}
}
