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

type UserHandler struct {
	*model.UserModel
}

type UserResponse struct {
	Data struct {
		Search struct {
			Edges []struct {
				Cursor string     `json:"cursor"`
				Node   model.User `json:"node"`
			} `json:"edges"`
			model.PageInfo `json:"pageInfo"`
		} `json:"search"`
		User struct {
			AvatarURL string          `json:"avatarUrl"`
			CreatedAt time.Time       `json:"createdAt"`
			Followers model.Directory `json:"followers"`
			Gists     struct {
				Edges []struct {
					Cursor string `json:"cursor"`
					Node   Gist   `json:"node"`
				} `json:"edges"`
				PageInfo   model.PageInfo `json:"pageInfo"`
				TotalCount int            `json:"totalCount"`
			} `json:"gists"`
			Location     string `json:"location"`
			Login        string `json:"login"`
			Name         string `json:"name"`
			Repositories struct {
				Edges []struct {
					Cursor string           `json:"cursor"`
					Node   model.Repository `json:"node"`
				} `json:"edges"`
				PageInfo   model.PageInfo `json:"pageInfo"`
				TotalCount int            `json:"totalCount"`
			} `json:"repositories"`
		} `json:"user"`
		model.RateLimit `json:"rateLimit"`
	} `json:"data"`
	Errors []model.Error `json:"errors"`
}

type UserRank struct {
	Login      string `bson:"_id"`
	TotalCount int    `bson:"total_count"`
}

type Gist struct {
	Name       string          `json:"name" bson:"name"`
	Stargazers model.Directory `json:"stargazers" bson:"stargazers"`
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		model.NewUserModel(),
	}
}

func (u *UserHandler) Init(starter chan<- struct{}) {
	logger.Info("Initializing user collection...")
	u.CreateIndexes()
	logger.Success("User collection initialized!")
	starter <- struct{}{}
}

func (u *UserHandler) Collect() error {
	logger.Info("Collecting users...")
	from := time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	q := model.Query{
		Schema: model.ReadQuery("users"),
		SearchArguments: model.SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}

	return u.Travel(&from, &q)
}

func (u *UserHandler) Travel(from *time.Time, q *model.Query) error {
	to := time.Now()
	if from.After(to) {
		return nil
	}

	q.SearchArguments.Query = strconv.Quote(util.ParseStruct(model.SearchQuery{
		Created:   fmt.Sprintf("%s..%s", from.Format(time.RFC3339), from.AddDate(0, 0, 7).Format(time.RFC3339)),
		Followers: ">=10",
		Repos:     ">=5",
		Sort:      "joined",
	}, " "))

	var users []model.User
	if err := u.FetchUsers(q, &users); err != nil {
		return err
	}
	u.StoreUsers(users)
	*from = from.AddDate(0, 0, 7)

	return u.Travel(from, q)
}

func (u *UserHandler) FetchUsers(q *model.Query, users *[]model.User) error {
	res := UserResponse{}
	if err := u.Fetch(*q, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		*users = append(*users, edge.Node)
	}
	res.Data.RateLimit.Break()
	if !res.Data.Search.PageInfo.HasNextPage {
		q.SearchArguments.After = ""
		return nil
	}
	q.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return u.FetchUsers(q, users)
}

func (u *UserHandler) StoreUsers(users []model.User) {
	if len(users) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var models []mongo.WriteModel
	for _, user := range users {
		filter := bson.D{{"_id", user.Login}}
		update := bson.D{{"$set", user}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}
	res, err := u.Model.Collection().BulkWrite(ctx, models)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if res.ModifiedCount > 0 {
		logger.Success(fmt.Sprintf("Updated %d users!", res.ModifiedCount))
	}
	if res.UpsertedCount > 0 {
		logger.Success(fmt.Sprintf("Inserted %d users!", res.UpsertedCount))
	}
}

func (u *UserHandler) Update() error {
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
	logger.Info("Updating user repositories...")
	reposQuery := model.NewReposQuery()
	for cursor.Next(ctx) {
		user := model.User{}
		if err := cursor.Decode(&user); err != nil {
			log.Fatalln(err.Error())
		}

		var gists []Gist
		gistsQuery.UserArguments.Login = strconv.Quote(user.Login)
		if err := u.FetchGists(gistsQuery, &gists); err != nil {
			return err
		}
		u.UpdateGists(user, gists)

		var repos []model.Repository
		reposQuery.UserArguments.Login = strconv.Quote(user.Login)
		if err := u.FetchRepositories(reposQuery, &repos); err != nil {
			return err
		}
		u.UpdateRepositories(user, repos)
	}

	return nil
}

func (u *UserHandler) FetchGists(q *model.Query, gists *[]Gist) error {
	res := UserResponse{}
	if err := u.Fetch(*q, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.User.Gists.Edges {
		*gists = append(*gists, edge.Node)
	}
	res.Data.RateLimit.Break()
	if !res.Data.User.Gists.PageInfo.HasNextPage {
		q.GistsArguments.After = ""
		return nil
	}
	q.GistsArguments.After = strconv.Quote(res.Data.User.Gists.PageInfo.EndCursor)

	return u.FetchGists(q, gists)
}

func (u *UserHandler) UpdateGists(user model.User, gists []Gist) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"_id", user.Login}}
	update := bson.D{{"$set", bson.D{{"gists", gists}}}}
	u.Model.Collection().FindOneAndUpdate(ctx, filter, update)
	logger.Success(fmt.Sprintf("Updated %d user gists!", len(gists)))
}

func (u *UserHandler) FetchRepositories(q *model.Query, repos *[]model.Repository) error {
	res := UserResponse{}
	if err := u.Fetch(*q, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.User.Repositories.Edges {
		*repos = append(*repos, edge.Node)
	}
	res.Data.RateLimit.Break()
	if !res.Data.User.Repositories.PageInfo.HasNextPage {
		q.RepositoriesArguments.After = ""
		return nil
	}
	q.RepositoriesArguments.After = strconv.Quote(res.Data.User.Repositories.PageInfo.EndCursor)

	return u.FetchRepositories(q, repos)
}

func (u *UserHandler) UpdateRepositories(user model.User, repos []model.Repository) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"_id", user.Login}}
	update := bson.D{{"$set", bson.D{{"repositories", repos}}}}
	u.Model.Collection().FindOneAndUpdate(ctx, filter, update)
	logger.Success(fmt.Sprintf("Updated %d user repositories!", len(repos)))
}

func (u *UserHandler) RankFollowers() {
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
	count := u.Rank(pipeline, field)
	logger.Success(fmt.Sprintf("Ranked %d user followers!", count))
}

func (u *UserHandler) RankGistStars() {
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
	count := u.Rank(pipeline, field)
	logger.Success(fmt.Sprintf("Ranked %d user gist stars!", count))
}

func (u *UserHandler) RankRepositoryStars() {
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
	count := u.Rank(pipeline, field)
	logger.Success(fmt.Sprintf("Ranked %d user repository stars!", count))
}

func (u *UserHandler) RankRepositoryStarsByLanguage() {
	logger.Info("Ranking user repository stars by language...")
	count := 0
	for _, language := range model.Languages() {
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
		count += u.Rank(pipeline, field)
	}
	logger.Success(fmt.Sprintf("Ranked %d user repository stars by language!", count))
}

func (u *UserHandler) Rank(pipeline []bson.D, field string) int {
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
		userRank := UserRank{}
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
			_, err := u.Model.Collection().BulkWrite(ctx, models)
			if err != nil {
				log.Fatalln(err.Error())
			}
			models = models[:0]
		}
	}

	return count
}

func (u *UserHandler) Fetch(q model.Query, res *UserResponse) (err error) {
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

func (u *UserHandler) GetByLogin(login string) (user model.User) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"_id", login}}
	if err := u.Model.Collection().FindOne(ctx, filter).Decode(&user); err != nil {
		logger.Warning(err.Error())
	}

	return user
}

func (u *UserHandler) CreateIndexes() {
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
