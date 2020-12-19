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

type OrganizationHandler struct {
	*model.OrganizationModel
}

func NewOrganizationHandler() *OrganizationHandler {
	return &OrganizationHandler{
		model.NewOrganizationModel(),
	}
}

func (o *OrganizationHandler) Init(starter chan<- struct{}) {
	logger.Info("Initializing organization collection...")
	o.CreateIndexes()
	logger.Success("Organization collection initialized!")
	starter <- struct{}{}
}

func (o *OrganizationHandler) Collect() error {
	logger.Info("Collecting organizations...")
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

func (o *OrganizationHandler) Travel(from *time.Time, q *model.Query) error {
	to := time.Now()
	if from.After(to) {
		return nil
	}

	q.SearchArguments.Query = strconv.Quote(util.ParseStruct(model.SearchQuery{
		Created: fmt.Sprintf("%s..%s", from.Format(time.RFC3339), from.AddDate(0, 0, 7).Format(time.RFC3339)),
		Repos:   ">=5",
		Sort:    "joined",
		Type:    "org",
	}, " "))

	var organizations []model.Organization
	if err := o.FetchOrganizations(q, &organizations); err != nil {
		return err
	}
	o.StoreOrganizations(organizations)
	*from = from.AddDate(0, 0, 7)

	return o.Travel(from, q)
}

func (o *OrganizationHandler) FetchOrganizations(q *model.Query, organizations *[]model.Organization) error {
	res := model.OrganizationResponse{}
	if err := o.Fetch(*q, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		*organizations = append(*organizations, edge.Node)
	}
	res.Data.RateLimit.Break()
	if !res.Data.Search.PageInfo.HasNextPage {
		q.SearchArguments.After = ""
		return nil
	}
	q.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return o.FetchOrganizations(q, organizations)
}

func (o *OrganizationHandler) StoreOrganizations(organizations []model.Organization) {
	if len(organizations) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var models []mongo.WriteModel
	for _, organization := range organizations {
		filter := bson.D{{"_id", organization.Login}}
		update := bson.D{{"$set", organization}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}
	res, err := o.Model.Collection().BulkWrite(ctx, models)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if res.ModifiedCount > 0 {
		logger.Success(fmt.Sprintf("Updated %d organizations!", res.ModifiedCount))
	}
	if res.UpsertedCount > 0 {
		logger.Success(fmt.Sprintf("Inserted %d organizations!", res.UpsertedCount))
	}
}

func (o *OrganizationHandler) Update() error {
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
	logger.Info("Updating organizations repositories...")
	repositoriesQuery := model.NewOrganizationRepositoriesQuery()
	for cursor.Next(ctx) {
		organization := model.Organization{}
		if err := cursor.Decode(&organization); err != nil {
			log.Fatalln(err.Error())
		}

		var repositories []model.Repository
		repositoriesQuery.UserArguments.Login = strconv.Quote(organization.Login)
		if err := o.FetchRepositories(repositoriesQuery, &repositories); err != nil {
			return err
		}
		o.UpdateRepositories(organization, repositories)
	}

	return nil
}

func (o *OrganizationHandler) FetchRepositories(q *model.Query, repositories *[]model.Repository) error {
	res := model.OrganizationResponse{}
	if err := o.Fetch(*q, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Organization.Repositories.Edges {
		*repositories = append(*repositories, edge.Node)
	}
	res.Data.RateLimit.Break()
	if !res.Data.Organization.Repositories.PageInfo.HasNextPage {
		q.RepositoriesArguments.After = ""
		return nil
	}
	q.RepositoriesArguments.After = strconv.Quote(res.Data.Organization.Repositories.PageInfo.EndCursor)

	return o.FetchRepositories(q, repositories)
}

func (o *OrganizationHandler) UpdateRepositories(organization model.Organization, repositories []model.Repository) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"_id", organization.Login}}
	update := bson.D{{"$set", bson.D{{"repositories", repositories}}}}
	o.Model.Collection().FindOneAndUpdate(ctx, filter, update)
	logger.Success(fmt.Sprintf("Updated %d organization repositories!", len(repositories)))
}

func (o *OrganizationHandler) Rank() {
	o.RankRepositoryStars()
	o.RankRepositoryStarsByLanguage()
}

func (o *OrganizationHandler) RankRepositoryStars() {
	logger.Info("Ranking organization repository stars...")
	pipeline := mongo.Pipeline{
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
	field := "ranks.repository_stars"
	count := o.Aggregate(pipeline, field)
	logger.Success(fmt.Sprintf("Ranked %d organization repository stars!", count))
}

func (o *OrganizationHandler) RankRepositoryStarsByLanguage() {
	logger.Info("Ranking organization repository stars by language...")
	count := 0
	for _, language := range util.Languages() {
		pipeline := mongo.Pipeline{
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
		field := fmt.Sprintf("ranks.repository_stars_%s", language)
		count += o.Aggregate(pipeline, field)
	}
	logger.Success(fmt.Sprintf("Ranked %d organization repository stars by language!", count))
}

func (o *OrganizationHandler) Aggregate(pipeline []bson.D, field string) int {
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
		userRank := model.UserRank{}
		if err := cursor.Decode(&userRank); err != nil {
			log.Fatalln(err.Error())
		}

		rank := model.Rank{
			Rank:       count + 1,
			TotalCount: userRank.TotalCount,
			CreatedAt:  time.Now(),
		}
		filter := bson.D{{"_id", userRank.Login}}
		update := bson.D{{"$set", bson.D{{field, rank}}}}
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

func (o *OrganizationHandler) Fetch(q model.Query, res *model.OrganizationResponse) (err error) {
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

func (o *OrganizationHandler) GetByLogin(login string) (organization model.Organization) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"_id", login}}
	if err := o.Model.Collection().FindOne(ctx, filter).Decode(&organization); err != nil {
		logger.Warning(err.Error())
	}

	return organization
}

func (o *OrganizationHandler) CreateIndexes() {
	if len(database.Indexes(o.Model.Name())) > 0 {
		return
	}

	database.CreateIndexes(o.Model.Name(), []string{
		"created_at",
		"name",
		"ranks.repository_stars.rank",
	})
}
