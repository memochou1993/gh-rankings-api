package model

import (
	"context"
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type RepositoryRank struct {
	NameWithOwner string `json:"nameWithOwner" bson:"nameWithOwner"`
	Rank          *Rank  `json:"rank,omitempty" bson:"rank,omitempty"`
}

type RepositoryRankModel struct {
	*Model
}

func (r *RepositoryRankModel) CreateIndexes() {
	database.CreateIndexes(r.Name(), []string{
		"nameWithOwner",
		"ranks.tags",
	})
}

func (r *RepositoryRankModel) Store(createdAt time.Time, p Pipeline) {
	ctx := context.Background()
	model := NewRepositoryModel()
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

		doc := RepositoryRank{
			NameWithOwner: rec.ID,
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
