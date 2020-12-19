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

func (u *OwnerHandler) Init(starter chan<- struct{}) {
	logger.Info("Initializing owner collection...")
	u.CreateIndexes()
	logger.Success("Owner collection initialized!")
	starter <- struct{}{}
}

func (u *OwnerHandler) Collect() error {
	logger.Info("Collecting owners...")
	from := time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	q := model.Query{
		Schema: model.ReadQuery("owners"),
		SearchArguments: model.SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}

	return u.Travel(&from, &q)
}

func (u *OwnerHandler) Travel(from *time.Time, q *model.Query) error {
	to := time.Now()
	if from.After(to) {
		return nil
	}

	q.SearchArguments.Query = strconv.Quote(util.ParseStruct(u.SearchQuery(*from), " "))

	var owners []model.Owner
	if err := u.FetchOwners(q, &owners); err != nil {
		return err
	}
	u.StoreOwners(owners)
	*from = from.AddDate(0, 0, 7)

	return u.Travel(from, q)
}

func (u *OwnerHandler) FetchOwners(q *model.Query, owners *[]model.Owner) error {
	res := model.OwnerResponse{}
	if err := u.Fetch(*q, &res); err != nil {
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

	return u.FetchOwners(q, owners)
}

func (u *OwnerHandler) StoreOwners(owners []model.Owner) {
	if len(owners) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var models []mongo.WriteModel
	for _, owner := range owners {
		filter := bson.D{{"_id", owner.Login}}
		update := bson.D{{"$set", owner}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}
	res, err := u.Model.Collection().BulkWrite(ctx, models)
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

func (u *OwnerHandler) Update() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := options.Find().SetBatchSize(1000)
	cursor, err := u.Model.Collection().Find(ctx, bson.D{}, opts)
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

		if u.IsUser(owner) {
			var gists []model.Gist
			gistsQuery.OwnerArguments.Login = strconv.Quote(owner.Login)
			if err := u.FetchGists(gistsQuery, &gists); err != nil {
				return err
			}
			u.UpdateGists(owner, gists)
		}

		var repositories []model.Repository
		repositoriesQuery.Field = u.Type(owner)
		repositoriesQuery.OwnerArguments.Login = strconv.Quote(owner.Login)
		if err := u.FetchRepositories(repositoriesQuery, &repositories); err != nil {
			return err
		}
		u.UpdateRepositories(owner, repositories)
	}

	return nil
}

func (u *OwnerHandler) FetchGists(q *model.Query, gists *[]model.Gist) error {
	res := model.OwnerResponse{}
	if err := u.Fetch(*q, &res); err != nil {
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

	return u.FetchGists(q, gists)
}

func (u *OwnerHandler) UpdateGists(owner model.Owner, gists []model.Gist) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"_id", owner.Login}}
	update := bson.D{{"$set", bson.D{{"gists", gists}}}}
	u.Model.Collection().FindOneAndUpdate(ctx, filter, update)
	logger.Success(fmt.Sprintf("Updated %d user gists!", len(gists)))
}

func (u *OwnerHandler) FetchRepositories(q *model.Query, repositories *[]model.Repository) error {
	res := model.OwnerResponse{}
	if err := u.Fetch(*q, &res); err != nil {
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

	return u.FetchRepositories(q, repositories)
}

func (u *OwnerHandler) UpdateRepositories(owner model.Owner, repositories []model.Repository) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"_id", owner.Login}}
	update := bson.D{{"$set", bson.D{{"repositories", repositories}}}}
	u.Model.Collection().FindOneAndUpdate(ctx, filter, update)
	logger.Success(fmt.Sprintf("Updated %d owner repositories!", len(repositories)))
}

func (u *OwnerHandler) Rank() {
	u.RankFollowers()
	u.RankGistStars()
	u.RankRepositoryStars()
	u.RankRepositoryStarsByLanguage()
}

func (u *OwnerHandler) RankFollowers() {
	logger.Info("Ranking user followers...")
	pipeline := mongo.Pipeline{
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
	field := "ranks.followers"
	count := u.Aggregate(pipeline, field)
	logger.Success(fmt.Sprintf("Ranked %d user followers!", count))
}

func (u *OwnerHandler) RankGistStars() {
	logger.Info("Ranking user gist stars...")
	pipeline := mongo.Pipeline{
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
	field := "ranks.gist_stars"
	count := u.Aggregate(pipeline, field)
	logger.Success(fmt.Sprintf("Ranked %d user gist stars!", count))
}

func (u *OwnerHandler) RankRepositoryStars() {
	logger.Info("Ranking user repository stars...")
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
	count := u.Aggregate(pipeline, field)
	logger.Success(fmt.Sprintf("Ranked %d user repository stars!", count))
}

func (u *OwnerHandler) RankRepositoryStarsByLanguage() {
	logger.Info("Ranking user repository stars by language...")
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
		count += u.Aggregate(pipeline, field)
	}
	logger.Success(fmt.Sprintf("Ranked %d user repository stars by language!", count))
}

func (u *OwnerHandler) Aggregate(pipeline []bson.D, field string) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := options.Aggregate().SetBatchSize(1000)
	cursor, err := u.Model.Collection().Aggregate(ctx, pipeline, opts)
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
		update := bson.D{{"$set", bson.D{{field, rank}}}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
		if cursor.RemainingBatchLength() == 0 {
			_, err := u.Model.Collection().BulkWrite(ctx, models)
			if err != nil {
				log.Fatalln(err.Error())
			}
			models = models[:0]
		}
	}

	return count
}

func (u *OwnerHandler) Fetch(q model.Query, res *model.OwnerResponse) (err error) {
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

func (u *OwnerHandler) GetByLogin(login string) (owner model.Owner) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"_id", login}}
	if err := u.Model.Collection().FindOne(ctx, filter).Decode(&owner); err != nil {
		logger.Warning(err.Error())
	}

	return owner
}

func (u *OwnerHandler) CreateIndexes() {
	if len(database.Indexes(u.Model.Name())) > 0 {
		return
	}

	database.CreateIndexes(u.Model.Name(), []string{
		"created_at",
		"name",
		"ranks.followers.rank",
		"ranks.gist_stars.rank",
		"ranks.repository_stars.rank",
	})
}

func (u *OwnerHandler) SearchQuery(from time.Time) model.SearchQuery {
	return model.SearchQuery{
		Created: fmt.Sprintf("%s..%s", from.Format(time.RFC3339), from.AddDate(0, 0, 7).Format(time.RFC3339)),
		Repos:   ">=5",
		Sort:    "joined",
	}
}

func (u *OwnerHandler) IsUser(owner model.Owner) bool {
	return owner.Followers != nil
}

func (u *OwnerHandler) IsOrganization(owner model.Owner) bool {
	return owner.Followers == nil
}

func (u *OwnerHandler) Type(owner model.Owner) (t string) {
	if u.IsUser(owner) {
		return "user"
	}
	if u.IsOrganization(owner) {
		return "organization"
	}
	return ""
}
