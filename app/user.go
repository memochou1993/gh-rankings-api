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
				AvatarURL    string    `json:"avatarUrl"`
				CreatedAt    time.Time `json:"createdAt"`
				Followers    List      `json:"followers"`
				Location     string    `json:"location"`
				Login        string    `json:"login"`
				Name         string    `json:"name"`
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
	AvatarURL    string       `json:"avatarUrl" bson:"avatar_url"`
	CreatedAt    time.Time    `json:"createdAt" bson:"created_at"`
	Followers    List         `json:"followers" bson:"followers"`
	Location     string       `json:"location" bson:"location"`
	Login        string       `json:"login" bson:"login"`
	Name         string       `json:"name" bson:"name"`
	Repositories []Repository `json:"repositories" bson:"repositories"`
	Ranks        struct {
		RepositoryStars struct {
			Rank       int       `bson:"rank"`
			TotalCount int       `bson:"total_count"`
			CreatedAt  time.Time `bson:"created_at"`
		} `bson:"repository_stars"`
	}
}

type Repository struct {
	Name            string `json:"name" bson:"name"`
	PrimaryLanguage struct {
		Name string `json:"name" bson:"name"`
	} `json:"primaryLanguage" bson:"primary_language"`
	Stargazers List `json:"stargazers" bson:"stargazers"`
}

type List struct {
	TotalCount int `json:"totalCount" bson:"total_count"`
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
	u.CreateIndexes()
	logger.Success("User collection initialized!")
	starter <- struct{}{}
}

func (u *UserCollection) Collect() error {
	logger.Info("Collecting users...")
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

	return u.Travel(&from, &q)
}

func (u *UserCollection) Travel(from *time.Time, q *Query) error {
	to := time.Now()
	if from.After(to) {
		logger.Warning("Take a vacation...")
		return nil
	}

	q.SearchArguments.Query = q.String(util.JoinStruct(SearchQuery{
		Created:   q.Range(*from, from.AddDate(0, 0, 6)),
		Followers: ">=10",
		Repos:     ">=5",
		Sort:      "joined",
	}, " "))

	var users []interface{}
	if err := u.FetchUsers(q, &users); err != nil {
		return err
	}
	u.StoreUsers(users)
	*from = from.AddDate(0, 0, 7)

	return u.Travel(from, q)
}

func (u *UserCollection) FetchUsers(q *Query, users *[]interface{}) error {
	if err := u.Fetch(q); err != nil {
		return err
	}
	if len(u.Response.Data.Search.Edges) == 0 {
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

func (u *UserCollection) StoreUsers(users []interface{}) {
	if len(users) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.InsertMany().SetOrdered(false)
	_, err := u.GetCollection().InsertMany(ctx, users, opts)

	count := len(users)
	if err, ok := err.(mongo.BulkWriteException); ok {
		for _, err := range err.WriteErrors {
			if err.Code != database.ErrorDuplicateKey {
				log.Fatalln(err.Error())
			}
			count--
			logger.Warning(err.WriteError.Error())
		}
	}
	logger.Success(fmt.Sprintf("Inserted %d users!", count))
}

func (u *UserCollection) Update() error {
	logger.Info("Updating users...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cursor, err := u.GetCollection().Find(ctx, bson.D{})
	if err != nil {
		log.Fatalln(err.Error())
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
	count := 0
	for ; cursor.Next(ctx); count++ {
		user := User{}
		if err := cursor.Decode(&user); err != nil {
			log.Fatalln(err.Error())
		}
		q.UserArguments.Login = q.String(user.Login)

		var repos []interface{}
		if err := u.FetchRepositories(&q, &repos); err != nil {
			return err
		}
		u.UpdateRepositories(user, repos)
	}
	logger.Success(fmt.Sprintf("Updated %d users!", count))

	return nil
}

func (u *UserCollection) FetchRepositories(q *Query, repos *[]interface{}) error {
	if err := u.Fetch(q); err != nil {
		return err
	}
	if len(u.Response.Data.User.Repositories.Edges) == 0 {
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

func (u *UserCollection) UpdateRepositories(user User, repos []interface{}) {
	if len(repos) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"login", user.Login}}
	update := bson.D{{"$set", bson.D{{"repositories", repos}}}}
	u.GetCollection().FindOneAndUpdate(ctx, filter, update)
	logger.Success(fmt.Sprintf("Updated %d user repositories!", len(repos)))
}

func (u *UserCollection) RankRepositoryStars() {
	logger.Info("Ranking user repository stars...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pipeline := mongo.Pipeline{
		bson.D{
			{"$project", bson.D{
				{"login", "$login"},
				{"ranks", bson.D{
					{"repository_stars", bson.D{
						{"total_count", bson.D{
							{"$sum", "$repositories.stargazers.total_count"},
						}},
					}},
				}},
			}},
		},
		bson.D{
			{"$sort", bson.D{
				{"ranks.repository_stars.total_count", -1},
			}},
		},
	}
	opts := options.Aggregate().SetBatchSize(1000)
	cursor, err := u.GetCollection().Aggregate(ctx, pipeline, opts)
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
		user := User{}
		user.Ranks.RepositoryStars.Rank = count + 1
		user.Ranks.RepositoryStars.CreatedAt = time.Now()
		if err := cursor.Decode(&user); err != nil {
			log.Fatalln(err.Error())
		}

		filter := bson.D{{"login", user.Login}}
		model := bson.D{{"$set", bson.D{{"ranks.repository_stars", user.Ranks.RepositoryStars}}}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(model))
		if cursor.RemainingBatchLength() == 0 {
			_, err := database.GetCollection("users").BulkWrite(ctx, models)
			if err != nil {
				log.Fatalln(err.Error())
			}
			models = models[:0]
		}
	}
	logger.Success(fmt.Sprintf("Ranked %d user repository stars!", count))
}

func (u *UserCollection) Fetch(q *Query) error {
	u.Response.Data.RateLimit.Break()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := Fetch(ctx, q, &u.Response); err != nil {
		return err
	}
	for _, err := range u.Response.Errors {
		return err
	}

	return nil
}

func (u *UserCollection) GetLast() (user User) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.FindOne().SetSort(bson.D{{"created_at", -1}})
	if err := u.GetCollection().FindOne(ctx, bson.D{}, opts).Decode(&user); err != nil {
		log.Fatalln(err.Error())
	}

	return user
}

func (u *UserCollection) CreateIndexes() {
	if len(database.GetIndexes(u.name)) > 0 {
		return
	}

	database.CreateIndexes(u.name, []string{
		"created_at",
		"ranks.repository_stars.rank",
	})
	database.CreateUniqueIndexes(u.name, []string{
		"login",
	})
}
