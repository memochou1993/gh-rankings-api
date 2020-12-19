package handler

import (
	"context"
	"fmt"
	"github.com/memochou1993/github-rankings/app"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
	"time"
)

type OwnerHandler struct {
	*model.OwnerModel
}

func NewOwnerHandler() *OwnerHandler {
	return &OwnerHandler{
		model.NewOwnerModel(),
	}
}

func (o *OwnerHandler) Init(starter chan<- struct{}) {
	logger.Info("Initializing owner collection...")
	o.CreateIndexes()
	logger.Success("Owner collection initialized!")
	starter <- struct{}{}
}

func (o *OwnerHandler) Collect() error {
	logger.Info("Collecting owners...")
	from := time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	q := model.Query{
		Schema: model.ReadQuery("owners"),
		SearchArguments: model.SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}

	return o.Travel(&from, &q)
}

func (o *OwnerHandler) Travel(from *time.Time, q *model.Query) error {
	to := time.Now()
	if from.After(to) {
		return nil
	}

	q.SearchArguments.Query = strconv.Quote(util.ParseStruct(o.SearchQuery(*from), " "))

	var owners []model.Owner
	if err := o.FetchOwners(q, &owners); err != nil {
		return err
	}
	o.StoreOwners(owners)
	*from = from.AddDate(0, 0, 7)

	return o.Travel(from, q)
}

func (o *OwnerHandler) FetchOwners(q *model.Query, owners *[]model.Owner) error {
	res := model.OwnerResponse{}
	if err := o.Fetch(*q, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		*owners = append(*owners, edge.Node)
	}
	res.Data.RateLimit.Break()
	if !res.Data.Search.PageInfo.HasNextPage {
		q.SearchArguments.After = ""
		return nil
	}
	q.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return o.FetchOwners(q, owners)
}

func (o *OwnerHandler) StoreOwners(owners []model.Owner) {
	if len(owners) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var models []mongo.WriteModel
	for _, owner := range owners {
		owner.Type = o.Type(owner)
		filter := bson.D{{"_id", owner.Login}}
		update := bson.D{{"$set", owner}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}
	res, err := o.Model.Collection().BulkWrite(ctx, models)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if res.ModifiedCount > 0 {
		logger.Success(fmt.Sprintf("Updated %d owners!", res.ModifiedCount))
	}
	if res.UpsertedCount > 0 {
		logger.Success(fmt.Sprintf("Inserted %d owners!", res.UpsertedCount))
	}
}

func (o *OwnerHandler) Update() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := options.Find().SetBatchSize(1000)
	cursor, err := o.Model.Collection().Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Fatalln(err.Error())
		}
	}()

	if cursor.RemainingBatchLength() == 0 {
		return nil
	}
	logger.Info("Updating user gists...")
	gistsQuery := model.NewGistsQuery()
	logger.Info("Updating owner repositories...")
	repositoriesQuery := model.NewOwnerRepositoriesQuery()
	for cursor.Next(ctx) {
		owner := model.Owner{}
		if err := cursor.Decode(&owner); err != nil {
			log.Fatalln(err.Error())
		}

		if o.Type(owner) == "user" {
			var gists []model.Gist
			gistsQuery.OwnerArguments.Login = strconv.Quote(owner.Login)
			if err := o.FetchGists(gistsQuery, &gists); err != nil {
				return err
			}
			o.UpdateGists(owner, gists)
		}

		var repositories []model.Repository
		repositoriesQuery.Field = o.Type(owner)
		repositoriesQuery.OwnerArguments.Login = strconv.Quote(owner.Login)
		if err := o.FetchRepositories(repositoriesQuery, &repositories); err != nil {
			return err
		}
		o.UpdateRepositories(owner, repositories)
	}

	return nil
}

func (o *OwnerHandler) FetchGists(q *model.Query, gists *[]model.Gist) error {
	res := model.OwnerResponse{}
	if err := o.Fetch(*q, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Owner.Gists.Edges {
		*gists = append(*gists, edge.Node)
	}
	res.Data.RateLimit.Break()
	if !res.Data.Owner.Gists.PageInfo.HasNextPage {
		q.GistsArguments.After = ""
		return nil
	}
	q.GistsArguments.After = strconv.Quote(res.Data.Owner.Gists.PageInfo.EndCursor)

	return o.FetchGists(q, gists)
}

func (o *OwnerHandler) UpdateGists(owner model.Owner, gists []model.Gist) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"_id", owner.Login}}
	update := bson.D{{"$set", bson.D{{"gists", gists}}}}
	o.Model.Collection().FindOneAndUpdate(ctx, filter, update)
	logger.Success(fmt.Sprintf("Updated %d user gists!", len(gists)))
}

func (o *OwnerHandler) FetchRepositories(q *model.Query, repositories *[]model.Repository) error {
	res := model.OwnerResponse{}
	if err := o.Fetch(*q, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Owner.Repositories.Edges {
		*repositories = append(*repositories, edge.Node)
	}
	res.Data.RateLimit.Break()
	if !res.Data.Owner.Repositories.PageInfo.HasNextPage {
		q.RepositoriesArguments.After = ""
		return nil
	}
	q.RepositoriesArguments.After = strconv.Quote(res.Data.Owner.Repositories.PageInfo.EndCursor)

	return o.FetchRepositories(q, repositories)
}

func (o *OwnerHandler) UpdateRepositories(owner model.Owner, repositories []model.Repository) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"_id", owner.Login}}
	update := bson.D{{"$set", bson.D{{"repositories", repositories}}}}
	o.Model.Collection().FindOneAndUpdate(ctx, filter, update)
	logger.Success(fmt.Sprintf("Updated %d %s repositories!", len(repositories), owner.Type))
}

func (o *OwnerHandler) Rank() {
	for _, t := range []string{"user"} {
		o.RankFollowers(t)
		o.RankGistStars(t)
	}
	for _, t := range []string{"user", "organization"} {
		o.RankRepositoryStars(t)
		o.RankRepositoryStarsByLanguage(t)
	}
}

func (o *OwnerHandler) RankFollowers(t string) {
	logger.Info("Ranking user followers...")
	pipeline := mongo.Pipeline{
		bson.D{
			{"$match", bson.D{
				{"type", t},
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
	}
	field := fmt.Sprintf("ranks.%s_followers", t)
	count := o.Aggregate(pipeline, field)
	logger.Success(fmt.Sprintf("Ranked %d user followers!", count))
}

func (o *OwnerHandler) RankGistStars(t string) {
	logger.Info("Ranking user gist stars...")
	pipeline := mongo.Pipeline{
		bson.D{
			{"$match", bson.D{
				{"type", t},
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
	}
	field := fmt.Sprintf("ranks.%s_gist_stars", t)
	count := o.Aggregate(pipeline, field)
	logger.Success(fmt.Sprintf("Ranked %d user gist stars!", count))
}

func (o *OwnerHandler) RankRepositoryStars(t string) {
	logger.Info(fmt.Sprintf("Ranking %s repository stars...", t))
	pipeline := mongo.Pipeline{
		bson.D{
			{"$match", bson.D{
				{"type", t},
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
	}
	field := fmt.Sprintf("ranks.%s_repository_stars", t)
	count := o.Aggregate(pipeline, field)
	logger.Success(fmt.Sprintf("Ranked %d %s repository stars!", count, t))
}

func (o *OwnerHandler) RankRepositoryStarsByLanguage(t string) {
	logger.Info(fmt.Sprintf("Ranking %s repository stars by language...", t))
	count := 0
	for _, language := range util.Languages() {
		pipeline := mongo.Pipeline{
			bson.D{
				{"$match", bson.D{
					{"type", t},
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
		}
		field := fmt.Sprintf("ranks.%s_repository_stars_%s", t, language)
		count += o.Aggregate(pipeline, field)
	}
	logger.Success(fmt.Sprintf("Ranked %d %s repository stars by language!", count, t))
}

func (o *OwnerHandler) Aggregate(pipeline []bson.D, field string) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := options.Aggregate().SetBatchSize(1000)
	cursor, err := o.Model.Collection().Aggregate(ctx, pipeline, opts)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Fatalln(err.Error())
		}
	}()

	var models []mongo.WriteModel
	count := 0
	for ; cursor.Next(ctx); count++ {
		ownerRank := model.OwnerRank{}
		if err := cursor.Decode(&ownerRank); err != nil {
			log.Fatalln(err.Error())
		}

		rank := model.Rank{
			Rank:       count + 1,
			TotalCount: ownerRank.TotalCount,
			CreatedAt:  time.Now(),
		}
		filter := bson.D{{"_id", ownerRank.Login}}
		update := bson.D{{"$push", bson.D{{field, rank}}}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
		if cursor.RemainingBatchLength() == 0 {
			_, err := o.Model.Collection().BulkWrite(ctx, models)
			if err != nil {
				log.Fatalln(err.Error())
			}
			models = models[:0]
		}
	}

	return count
}

func (o *OwnerHandler) Fetch(q model.Query, res *model.OwnerResponse) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Fetch(ctx, fmt.Sprint(q), res); err != nil {
		return err
	}
	for _, err := range res.Errors {
		return err
	}

	return nil
}

func (o *OwnerHandler) GetByLogin(login string) (owner model.Owner) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"_id", login}}
	if err := o.Model.Collection().FindOne(ctx, filter).Decode(&owner); err != nil {
		logger.Warning(err.Error())
	}

	return owner
}

func (o *OwnerHandler) CreateIndexes() {
	if len(database.Indexes(o.Model.Name())) > 0 {
		return
	}

	database.CreateIndexes(o.Model.Name(), []string{
		"created_at",
		"name",
		"user_followers",
		"user_gist_stars",
		"user_repository_stars",
		"organization_repository_stars",
	})
}

func (o *OwnerHandler) SearchQuery(from time.Time) model.SearchQuery {
	return model.SearchQuery{
		Created: fmt.Sprintf("%s..%s", from.Format(time.RFC3339), from.AddDate(0, 0, 7).Format(time.RFC3339)),
		Repos:   ">=5",
		Sort:    "joined",
	}
}

func (o *OwnerHandler) Type(owner model.Owner) (t string) {
	t = "user"
	if owner.Followers == nil {
		t = "organization"
	}
	return t
}
