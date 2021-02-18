package pipeline

import (
	"fmt"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/pipeline/operator"
	"github.com/memochou1993/gh-rankings/app/resource"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	ascending  = 1
	descending = -1
)

type Pipeline struct {
	Pipeline *mongo.Pipeline
	Type     string
	Field    string
	Language string
	Location string
}

func User() (pipelines []*Pipeline) {
	rankType := app.TypeUser
	fields := []string{
		"followers",
		"gists.forks",
		"gists.stargazers",
		"repositories.forks",
		"repositories.stargazers",
		"repositories.watchers",
	}
	for _, field := range fields {
		pipelines = append(pipelines, rankByField(rankType, field))
		pipelines = append(pipelines, rankByLocation(rankType, field)...)
	}
	pipelines = append(pipelines, rankOwnerRepositoryByLanguage(rankType, "repositories.stargazers")...)
	pipelines = append(pipelines, rankOwnerRepositoryByLanguage(rankType, "repositories.forks")...)
	pipelines = append(pipelines, rankOwnerRepositoryByLanguage(rankType, "repositories.watchers")...)
	return
}

func Organization() (pipelines []*Pipeline) {
	rankType := app.TypeOrganization
	fields := []string{
		"repositories.forks",
		"repositories.stargazers",
		"repositories.watchers",
	}
	for _, field := range fields {
		pipelines = append(pipelines, rankByField(rankType, field))
		pipelines = append(pipelines, rankByLocation(rankType, field)...)
		pipelines = append(pipelines, rankOwnerRepositoryByLanguage(rankType, field)...)
	}
	return
}

func Repository() (pipelines []*Pipeline) {
	rankType := app.TypeRepository
	fields := []string{
		"forks",
		"stargazers",
		"watchers",
	}
	for _, field := range fields {
		pipelines = append(pipelines, rankByField(rankType, field))
		pipelines = append(pipelines, rankRepositoryByLanguage(rankType, field)...)
	}
	return
}

func RankCount(pipeline mongo.Pipeline) mongo.Pipeline {
	stages := mongo.Pipeline{
		operator.Match("total_count", bson.D{{"$gt", 0}}),
		operator.Count(),
	}
	return append(pipeline, stages...)
}

func Search(req *request.Request) mongo.Pipeline {
	cond := mongo.Pipeline{{
		{"type", req.Type},
		{"field", req.Field},
		{"language", req.Language},
		{"location", req.Location},
		{"created_at", operator.In(req.Timestamps)},
	}}
	if req.Name != "" {
		cond = append(cond, bson.D{{"name", req.Name}})
	}
	return mongo.Pipeline{
		operator.Match("$and", cond),
		operator.Skip((req.Page - 1) * req.Limit),
		operator.Limit(req.Limit),
	}
}

func List(req *request.Request) mongo.Pipeline {
	cond := mongo.Pipeline{{
		{"created_at", operator.In(req.Timestamps)},
	}}
	if req.Name != "" {
		cond = append(cond, bson.D{{"name", req.Name}})
	}
	if req.Type != "" {
		cond = append(cond, bson.D{{"type", req.Type}})
	}
	if req.Field != "" {
		cond = append(cond, bson.D{{"field", req.Field}})
	}
	if req.Language != "" {
		cond = append(cond, bson.D{{"language", req.Language}})
	}
	if req.Location != "" {
		cond = append(cond, bson.D{{"location", req.Location}})
	}
	return mongo.Pipeline{
		operator.Match("$and", cond),
		operator.Skip((req.Page - 1) * req.Limit),
		operator.Limit(req.Limit),
	}
}

func rankByField(rankType string, field string) *Pipeline {
	return &Pipeline{
		Pipeline: &mongo.Pipeline{
			operator.Project(bson.D{
				id(),
				imageUrl(),
				totalCount(field),
			}),
			operator.Sort("total_count", descending),
		},
		Type:  rankType,
		Field: field,
	}
}

func rankByLocation(rankType string, field string) (pipelines []*Pipeline) {
	for _, location := range resource.Locations {
		pipelines = append(pipelines, &Pipeline{
			Pipeline: &mongo.Pipeline{
				operator.Match("parsed_location", location.Name),
				operator.Project(bson.D{
					id(),
					imageUrl(),
					totalCount(field),
				}),
				operator.Sort("total_count", descending),
			},
			Type:     rankType,
			Field:    field,
			Location: location.Name,
		})
		for _, city := range location.Cities {
			location := fmt.Sprintf("%s, %s", city.Name, location.Name)
			pipelines = append(pipelines, &Pipeline{
				Pipeline: &mongo.Pipeline{
					operator.Match("parsed_city", location),
					operator.Project(bson.D{
						id(),
						imageUrl(),
						totalCount(field),
					}),
					operator.Sort("total_count", descending),
				},
				Type:     rankType,
				Field:    field,
				Location: location,
			})
		}
	}
	return
}

func rankOwnerRepositoryByLanguage(rankType string, field string) (pipelines []*Pipeline) {
	for _, language := range resource.Languages {
		pipelines = append(pipelines, &Pipeline{
			Pipeline: &mongo.Pipeline{
				operator.Unwind("repositories"),
				operator.Match("repositories.primary_language.name", language.Name),
				operator.Group(bson.D{
					id(),
					{"image_url", operator.First("$image_url")},
					totalCount(field),
				}),
				operator.Sort("total_count", descending),
			},
			Type:     rankType,
			Field:    field,
			Language: language.Name,
		})
	}
	return
}

func rankRepositoryByLanguage(rankType string, field string) (pipelines []*Pipeline) {
	for _, language := range resource.Languages {
		pipelines = append(pipelines, &Pipeline{
			Pipeline: &mongo.Pipeline{
				operator.Match("primary_language.name", language.Name),
				operator.Project(bson.D{
					id(),
					imageUrl(),
					totalCount(field),
				}),
				operator.Sort("total_count", descending),
			},
			Type:     rankType,
			Field:    field,
			Language: language.Name,
		})
	}
	return
}

func id() bson.E {
	return bson.E{
		Key:   "_id",
		Value: "$_id",
	}
}

func imageUrl() bson.E {
	return bson.E{
		Key:   "image_url",
		Value: "$image_url",
	}
}

func totalCount(field string) bson.E {
	return bson.E{
		Key:   "total_count",
		Value: operator.Sum(fmt.Sprintf("%s.total_count", field)),
	}
}
