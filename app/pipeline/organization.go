package pipeline

import (
	"fmt"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/pipeline/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RankOrganization() (pipelines []*Pipeline) {
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

func SearchOrganizations(req *request.Organization) mongo.Pipeline {
	cond := mongo.Pipeline{}
	if req.Q != "" {
		cond = append(cond, bson.D{{"_id", operator.Regex(fmt.Sprintf(".*%s.*", req.Q), "i")}})
	}
	return mongo.Pipeline{
		operator.Match("$or", cond),
		operator.Project(bson.D{{"repositories", 0}}),
		operator.Skip((req.Page - 1) * req.Limit),
		operator.Limit(req.Limit),
	}
}

func ListOrganizations(req *request.Organization) mongo.Pipeline {
	return mongo.Pipeline{
		operator.Project(bson.D{{"repositories", 0}}),
		operator.Skip((req.Page - 1) * req.Limit),
		operator.Limit(req.Limit),
	}
}
