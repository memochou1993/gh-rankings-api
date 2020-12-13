package app

import (
	"context"
	"fmt"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type UserCollection struct {
	Collection
	Response struct {
		Data struct {
			Search struct {
				UserCount int `json:"userCount"`
				Edges     []struct {
					Cursor string `json:"cursor"`
					Node   User   `json:"node"`
				} `json:"edges"`
				PageInfo PageInfo `json:"pageInfo"`
			} `json:"search"`
			User struct {
				AvatarURL string    `json:"avatarUrl"`
				CreatedAt time.Time `json:"createdAt"`
				Email     string    `json:"email"`
				Followers struct {
					TotalCount int `json:"totalCount"`
				} `json:"followers"`
				Location     string `json:"location"`
				Login        string `json:"login"`
				Name         string `json:"name"`
				Repositories struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Cursor string     `json:"cursor"`
						Node   Repository `json:"node"`
					} `json:"edges"`
					PageInfo PageInfo `json:"pageInfo"`
				} `json:"repositories"`
			} `json:"user"`
			RateLimit RateLimit `json:"rateLimit"`
		} `json:"data"`
		Errors []Error `json:"errors"`
	}
}

type User struct {
	AvatarURL string    `json:"avatarUrl" bson:"avatar_url"`
	CreatedAt time.Time `json:"createdAt" bson:"created_at"`
	Email     string    `json:"email" bson:"email"`
	Followers struct {
		TotalCount int `json:"totalCount" bson:"total_count"`
	} `json:"followers" bson:"followers"`
	Location     string       `json:"location" bson:"location"`
	Login        string       `json:"login" bson:"login"`
	Name         string       `json:"name" bson:"name"`
	Repositories []Repository `json:"repositories" bson:"repositories"`
}

type Repository struct {
	Name            string `json:"name" bson:"name"`
	PrimaryLanguage struct {
		Name string `json:"name" bson:"name"`
	} `json:"primaryLanguage" bson:"primary_language"`
	Stargazers struct {
		TotalCount int `json:"totalCount" bson:"total_count"`
	} `json:"stargazers" bson:"stargazers"`
}

// TODO: add to user struct
type UserRanking struct {
	Login             string            `bson:"login"`
	RepositoryRanking RepositoryRanking `bson:"ranking"`
}

type RepositoryRanking struct {
	Rank            int
	RepositoryStars int       `bson:"repository_stars"`
	CreatedAt       time.Time `bson:"created_at"`
}

func NewUserCollection() *UserCollection {
	return &UserCollection{
		Collection: Collection{
			name: "users",
		},
	}
}

func (u *UserCollection) Init(starter chan<- struct{}) {
	logger.Info("Initializing user collection...")
	if err := u.Index(); err != nil {
		log.Fatalln(err.Error())
	}
	logger.Success("User collection initialized...")
	starter <- struct{}{}
}

func (u *UserCollection) Collect() error {
	from := time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	if database.Count(u.name) > 0 {
		from = u.GetLast().CreatedAt.Truncate(24 * time.Hour)
	}
	q := Query{
		Schema: ReadQuery("users"),
		SearchArguments: SearchArguments{
			First: 100,
			Type:  "USER",
		},
	}
	if err := u.Travel(&from, &q); err != nil {
		return err
	}

	return nil
}

func (u *UserCollection) Travel(from *time.Time, q *Query) error {
	to := time.Now()
	if from.After(to) {
		logger.Warning("Take a break...")
		time.Sleep(7 * 24 * time.Hour)

		return nil
	}

	q.SearchArguments.Query = q.String(util.JoinStruct(SearchQuery{
		Created:   q.Range(*from, from.AddDate(0, 0, 6)),
		Followers: ">=10",
		Repos:     ">=5",
	}, " "))

	var users []interface{}
	if err := u.FetchUsers(q, &users); err != nil {
		return err
	}
	if err := u.StoreUsers(users); err != nil {
		return err
	}
	*from = from.AddDate(0, 0, 7)

	return u.Travel(from, q)
}

func (u *UserCollection) FetchUsers(q *Query, users *[]interface{}) error {
	logger.Debug(q.SearchArguments)
	logger.Info("Searching users...")
	if err := u.Fetch(q); err != nil {
		return err
	}
	count := len(u.Response.Data.Search.Edges)
	logger.Success(fmt.Sprintf("%d users discovered", count))
	if count == 0 {
		return nil
	}
	for _, edge := range u.Response.Data.Search.Edges {
		*users = append(*users, edge.Node)
	}

	if !u.Response.Data.Search.PageInfo.HasNextPage {
		q.SearchArguments.After = ""
		return nil
	}
	q.SearchArguments.After = q.String(u.Response.Data.Search.PageInfo.EndCursor)

	return u.FetchUsers(q, users)
}

func (u *UserCollection) StoreUsers(users []interface{}) error {
	if len(users) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.InsertMany().SetOrdered(false)
	_, err := u.GetCollection().InsertMany(ctx, users, opts)

	count := len(users)
	if err, ok := err.(mongo.BulkWriteException); ok {
		for _, err := range err.WriteErrors {
			if err.Code != 11000 {
				return err
			}
			count--
			logger.Warning(err.Message)
		}
	}
	logger.Success(fmt.Sprintf("%d users inserted", count))

	return nil
}

func (u *UserCollection) Update() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cursor, err := u.GetCollection().Find(ctx, bson.D{})
	if err != nil {
		return err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Fatalln(err.Error())
		}
	}()

	q := Query{
		Schema: ReadQuery("user_repositories"),
		RepositoriesArguments: RepositoriesArguments{
			First:             100,
			OrderBy:           "{field:STARGAZERS,direction:DESC}",
			OwnerAffiliations: "OWNER",
		},
	}
	for cursor.Next(ctx) {
		user := User{}
		if err := cursor.Decode(&user); err != nil {
			return nil
		}
		q.UserArguments.Login = q.String(user.Login)

		var repos []interface{}
		if err := u.FetchRepositories(&q, &repos); err != nil {
			return err
		}
		u.StoreRepositories(user, repos)
	}

	return nil
}

func (u *UserCollection) FetchRepositories(q *Query, repos *[]interface{}) error {
	logger.Debug(q.UserArguments)
	logger.Debug(q.RepositoriesArguments)
	logger.Info("Searching user repositories...")
	if err := u.Fetch(q); err != nil {
		return err
	}
	count := len(u.Response.Data.User.Repositories.Edges)
	logger.Success(fmt.Sprintf("%d user repositories discovered", count))
	if count == 0 {
		return nil
	}
	for _, edge := range u.Response.Data.User.Repositories.Edges {
		*repos = append(*repos, edge.Node)
	}

	if !u.Response.Data.User.Repositories.PageInfo.HasNextPage {
		q.RepositoriesArguments.After = ""
		return nil
	}
	q.RepositoriesArguments.After = q.String(u.Response.Data.User.Repositories.PageInfo.EndCursor)

	return u.FetchRepositories(q, repos)
}

func (u *UserCollection) StoreRepositories(user User, repos []interface{}) {
	if len(repos) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"login", user.Login}}
	update := bson.D{{"$set", bson.D{{"repositories", repos}}}}

	u.GetCollection().FindOneAndUpdate(ctx, filter, update)
	logger.Success(fmt.Sprintf("%d user repositories inserted", len(repos)))
}

func (u *UserCollection) RankRepositoryStars() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	pipeline := mongo.Pipeline{
		bson.D{
			{"$project", bson.D{
				{"login", "$login"},
				{"ranking", bson.D{
					{"repository_stars", bson.D{{"$sum", "$repositories.stargazers.total_count"}}},
				}},
			}},
		},
		bson.D{
			{"$sort", bson.D{
				{"ranking.repository_stars", -1},
			}},
		},
	}

	cursor, err := u.GetCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Fatalln(err.Error())
		}
	}()

	for i := 1; cursor.Next(ctx); i++ {
		userRanking := UserRanking{
			RepositoryRanking: RepositoryRanking{
				Rank:      i,
				CreatedAt: time.Now(),
			},
		}
		if err := cursor.Decode(&userRanking); err != nil {
			return err
		}

		filter := bson.D{{"login", userRanking.Login}}
		update := bson.D{{"$push", bson.D{{"rankings.repository_stars", userRanking.RepositoryRanking}}}}
		opts := options.Update().SetUpsert(true)
		_, err := database.GetCollection("users").UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *UserCollection) Fetch(q *Query) error {
	u.Response.Data.RateLimit.Check()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := Fetch(ctx, q, &u.Response)
	logger.Debug(u.Response.Data.RateLimit)
	for _, err := range u.Response.Errors {
		return err
	}

	return err
}

func (u *UserCollection) GetLast() (user User) {
	opts := options.FindOne().SetSort(bson.D{{"created_at", -1}})
	if err := database.Get(u.name, opts).Decode(&user); err != nil {
		log.Fatalln(err.Error())
	}

	return user
}

func (u *UserCollection) Index() error {
	if len(database.GetIndexes(u.name)) > 0 {
		return nil
	}

	indexes := []string{"created_at"}
	if err := database.CreateIndexes(u.name, indexes); err != nil {
		return err
	}

	uniqueIndexes := []string{"login"}
	if err := database.CreateUniqueIndexes(u.name, uniqueIndexes); err != nil {
		return err
	}

	return nil
}
