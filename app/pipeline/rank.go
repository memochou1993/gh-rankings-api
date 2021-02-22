package pipeline

import (
	"fmt"
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/pipeline/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func SearchRanks(req *request.Rank) mongo.Pipeline {
	cond := mongo.Pipeline{{
		{"type", req.Type},
		{"field", req.Field},
		{"language", req.Language},
		{"location", req.Location},
		{"created_at", operator.In(req.Timestamps)},
	}}
	if req.Name != "" {
		cond = append(cond, bson.D{{"name", operator.Regex(fmt.Sprintf(".*%s.*", req.Name), "i")}})
	}
	return mongo.Pipeline{
		operator.Match("$and", cond),
		operator.Skip((req.Page - 1) * req.Limit),
		operator.Limit(req.Limit),
	}
}

func ListRanks(req *request.Rank) mongo.Pipeline {
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
