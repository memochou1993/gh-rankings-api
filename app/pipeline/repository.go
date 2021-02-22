package pipeline

import (
	"fmt"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/pipeline/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RankRepository() (pipelines []*Pipeline) {
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

func SearchRepositories(req *request.Repository) mongo.Pipeline {
	cond := mongo.Pipeline{}
	if req.Q != "" {
		cond = append(cond, bson.D{{"_id", operator.Regex(fmt.Sprintf(".*%s.*", req.Q), "i")}})
	}
	return mongo.Pipeline{
		operator.Match("$or", cond),
		operator.Skip((req.Page - 1) * req.Limit),
		operator.Limit(req.Limit),
	}
}

func ListRepositories(req *request.Repository) mongo.Pipeline {
	return mongo.Pipeline{
		operator.Skip((req.Page - 1) * req.Limit),
		operator.Limit(req.Limit),
	}
}
