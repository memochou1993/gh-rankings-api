package pipeline

import (
	"fmt"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/resource"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Pipeline struct {
	Pipeline *mongo.Pipeline
	Type     string
	Field    string
	Language string
	Location string
}

func RankByField(rankType string, field string) *Pipeline {
	return &Pipeline{
		Pipeline: &mongo.Pipeline{
			stageProject(field),
			stageSort(),
		},
		Type:  rankType,
		Field: field,
	}
}

func RankByLocation(rankType string, field string) (pipelines []*Pipeline) {
	for _, location := range resource.Locations {
		pipelines = append(pipelines, &Pipeline{
			Pipeline: &mongo.Pipeline{
				stageMatch("parsed_location", location.Name),
				stageProject(field),
				stageSort(),
			},
			Type:     rankType,
			Field:    field,
			Location: location.Name,
		})
		for _, city := range location.Cities {
			location := fmt.Sprintf("%s, %s", city.Name, location.Name)
			pipelines = append(pipelines, &Pipeline{
				Pipeline: &mongo.Pipeline{
					stageMatch("parsed_city", location),
					stageProject(field),
					stageSort(),
				},
				Type:     rankType,
				Field:    field,
				Location: location,
			})
		}
	}
	return
}

func RankOwnerRepositoryByLanguage(rankType string, field string) (pipelines []*Pipeline) {
	for _, language := range resource.Languages {
		pipelines = append(pipelines, &Pipeline{
			Pipeline: &mongo.Pipeline{
				stageUnwind("repositories"),
				stageMatch("repositories.primary_language.name", language.Name),
				stageGroup(field),
				stageSort(),
			},
			Type:     rankType,
			Field:    field,
			Language: language.Name,
		})
	}
	return
}

func RankRepositoryByLanguage(rankType string, field string) (pipelines []*Pipeline) {
	for _, language := range resource.Languages {
		pipelines = append(pipelines, &Pipeline{
			Pipeline: &mongo.Pipeline{
				stageMatch("primary_language.name", language.Name),
				stageProject(field),
				stageSort(),
			},
			Type:     rankType,
			Field:    field,
			Language: language.Name,
		})
	}
	return
}

func RankCount(p mongo.Pipeline) mongo.Pipeline {
	stages := []bson.D{
		stageMatch("total_count", bson.D{{"$gt", 0}}),
		stageCount(),
	}
	return append(p, stages...)
}

func List(req *request.Request, createdAt time.Time) mongo.Pipeline {
	return mongo.Pipeline{
		stageMatch("$and", mongo.Pipeline{{
			{"type", req.Type},
			{"field", req.Field},
			{"language", req.Language},
			{"location", req.Location},
			{"created_at", createdAt},
		}}),
		stageSkip((req.Page - 1) * req.Limit),
		stageLimit(req.Limit),
	}
}

func stageUnwind(field string) bson.D {
	return bson.D{
		{"$unwind", fmt.Sprintf("$%s", field)},
	}
}

func stageMatch(key string, value interface{}) bson.D {
	return bson.D{
		{"$match", bson.D{
			{key, value},
		}},
	}
}

// FIXME: should rename
func stageProject(field string) bson.D {
	return bson.D{
		{"$project", bson.D{
			{"_id", "$_id"},
			{"image_url", "$image_url"},
			{"total_count", bson.D{
				{"$sum", fmt.Sprintf("$%s.total_count", field)},
			}},
		}},
	}
}

// FIXME: should rename
func stageGroup(field string) bson.D {
	return bson.D{
		{"$group", bson.D{
			{"_id", "$_id"},
			{"image_url", bson.D{
				{"$first", "$image_url"},
			}},
			{"total_count", bson.D{
				{"$sum", fmt.Sprintf("$%s.total_count", field)},
			}},
		}},
	}
}

func stageSort() bson.D {
	return bson.D{
		{"$sort", bson.D{
			{"total_count", -1},
		}},
	}
}

func stageCount() bson.D {
	return bson.D{
		{"$count", "count"},
	}
}

func stageSkip(skip int64) bson.D {
	return bson.D{
		{"$skip", skip},
	}
}

func stageLimit(limit int64) bson.D {
	return bson.D{
		{"$limit", limit},
	}
}
