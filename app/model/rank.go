package model

import (
	"github.com/memochou1993/github-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Rank struct {
	Rank       int       `bson:"rank"`
	TotalCount int       `bson:"total_count"`
	Tags       []string  `bson:"tags"`
	Batch      int       `bson:"batch"`
	CreatedAt  time.Time `bson:"created_at"`
}

type RankPipeline struct {
	Pipeline mongo.Pipeline
	Tags     []string
}

func FollowersPipeline(ownerType string) RankPipeline {
	return RankPipeline{
		Pipeline: mongo.Pipeline{
			bson.D{
				{"$match", bson.D{
					{"type", ownerType},
				}},
			},
			bson.D{
				{"$project", bson.D{
					{"_id", "$_id"},
					{"total_count", bson.D{
						{"$sum", "$followers.total_count"},
					}},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"total_count", -1},
				}},
			},
		},
		Tags: []string{ownerType, "followers"},
	}
}

func GistForksPipeline(ownerType string) RankPipeline {
	return RankPipeline{
		Pipeline: mongo.Pipeline{
			bson.D{
				{"$match", bson.D{
					{"type", ownerType},
				}},
			},
			bson.D{
				{"$project", bson.D{
					{"_id", "$_id"},
					{"total_count", bson.D{
						{"$sum", "$gists.forks.total_count"},
					}},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"total_count", -1},
				}},
			},
		},
		Tags: []string{ownerType, "gist", "forks"},
	}
}

func GistStarsPipeline(ownerType string) RankPipeline {
	return RankPipeline{
		Pipeline: mongo.Pipeline{
			bson.D{
				{"$match", bson.D{
					{"type", ownerType},
				}},
			},
			bson.D{
				{"$project", bson.D{
					{"_id", "$_id"},
					{"total_count", bson.D{
						{"$sum", "$gists.stargazers.total_count"},
					}},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"total_count", -1},
				}},
			},
		},
		Tags: []string{ownerType, "gist", "stars"},
	}
}

func RepositoryForksPipeline(ownerType string) RankPipeline {
	return RankPipeline{
		Pipeline: mongo.Pipeline{
			bson.D{
				{"$match", bson.D{
					{"type", ownerType},
				}},
			},
			bson.D{
				{"$project", bson.D{
					{"_id", "$_id"},
					{"total_count", bson.D{
						{"$sum", "$repositories.forks.total_count"},
					}},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"total_count", -1},
				}},
			},
		},
		Tags: []string{ownerType, "repository", "forks"},
	}
}

func RepositoryStarsPipeline(ownerType string) RankPipeline {
	return RankPipeline{
		Pipeline: mongo.Pipeline{
			bson.D{
				{"$match", bson.D{
					{"type", ownerType},
				}},
			},
			bson.D{
				{"$project", bson.D{
					{"_id", "$_id"},
					{"total_count", bson.D{
						{"$sum", "$repositories.stargazers.total_count"},
					}},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"total_count", -1},
				}},
			},
		},
		Tags: []string{ownerType, "repository", "stars"},
	}
}

func RepositoryStarsPipelinesByLanguage(ownerType string) (pipelines []RankPipeline) {
	for _, language := range util.Languages() {
		pipelines = append(pipelines, RankPipeline{
			Pipeline: mongo.Pipeline{
				bson.D{
					{"$match", bson.D{
						{"type", ownerType},
					}},
				},
				bson.D{
					{"$unwind", "$repositories"},
				},
				bson.D{
					{"$match", bson.D{
						{"repositories.primary_language.name", language},
					}},
				},
				bson.D{
					{"$group", bson.D{
						{"_id", "$_id"},
						{"total_count", bson.D{
							{"$sum", "$repositories.stargazers.total_count"},
						}},
					}},
				},
				bson.D{
					{"$sort", bson.D{
						{"total_count", -1},
					}},
				},
			},
			Tags: []string{ownerType, "repository", "stars", language},
		})
	}
	return
}
