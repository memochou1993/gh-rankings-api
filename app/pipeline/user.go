package pipeline

import (
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/pipeline/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RankUser() (pipelines []*Pipeline) {
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

func SearchUsers(req *request.User) mongo.Pipeline {
	cond := mongo.Pipeline{}
	if req.Q != "" {
		cond = append(cond, bson.D{{"_id", operator.Regex(req.Q, "i")}})
	}
	return mongo.Pipeline{
		operator.Match("$or", cond),
		operator.Project(bson.D{{"repositories", 0}, {"gists", 0}}),
		operator.Skip((req.Page - 1) * req.Limit),
		operator.Limit(req.Limit),
	}
}

func ListUsers(req *request.User) mongo.Pipeline {
	return mongo.Pipeline{
		operator.Project(bson.D{{"repositories", 0}, {"gists", 0}}),
		operator.Skip((req.Page - 1) * req.Limit),
		operator.Limit(req.Limit),
	}
}
