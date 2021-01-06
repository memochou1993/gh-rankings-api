package model

import (
	"context"
	"github.com/memochou1993/github-rankings/app/resource"
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type Owner struct {
	AvatarURL    string       `json:"avatarUrl" bson:"avatar_url"`
	CreatedAt    time.Time    `json:"createdAt" bson:"created_at"`
	Followers    *Directory   `json:"followers" bson:"followers"`
	Location     string       `json:"location" bson:"location"`
	Login        string       `json:"login" bson:"_id"`
	Name         string       `json:"name" bson:"name"`
	Gists        []Gist       `json:"gists,omitempty" bson:"gists,omitempty"`
	Repositories []Repository `json:"repositories,omitempty" bson:"repositories,omitempty"`
	Tags         []string     `json:"tags" bson:"tags"`
	Ranks        []Rank       `json:"ranks" bson:"ranks,omitempty"`
}

func (o *Owner) ID() string {
	return o.Login
}

func (o *Owner) HasFollowers() bool {
	return o.Followers != nil
}

func (o *Owner) IsUser() bool {
	return o.HasFollowers()
}

func (o *Owner) IsOrganization() bool {
	return !o.HasFollowers()
}

func (o *Owner) TagType() {
	o.Tags = append(o.Tags, o.Type())
}

func (o *Owner) TagLocations() {
	o.Tags = append(o.Tags, resource.Locate(o.Location)...)
}

func (o *Owner) Type() (t string) {
	t = TypeUser
	if o.IsOrganization() {
		t = TypeOrganization
	}
	return
}

type OwnerResponse struct {
	Data struct {
		Search struct {
			Edges []struct {
				Cursor string `json:"cursor"`
				Node   Owner  `json:"node"`
			} `json:"edges"`
			PageInfo `json:"pageInfo"`
		} `json:"search"`
		Owner struct {
			AvatarURL string    `json:"avatarUrl"`
			CreatedAt time.Time `json:"createdAt"`
			Followers Directory `json:"followers"`
			Gists     struct {
				Edges []struct {
					Cursor string `json:"cursor"`
					Node   Gist   `json:"node"`
				} `json:"edges"`
				PageInfo `json:"pageInfo"`
			} `json:"gists"`
			Location     string `json:"location"`
			Login        string `json:"login"`
			Name         string `json:"name"`
			Repositories struct {
				Edges []struct {
					Cursor string     `json:"cursor"`
					Node   Repository `json:"node"`
				} `json:"edges"`
				PageInfo `json:"pageInfo"`
			} `json:"repositories"`
		} `json:"owner"`
		RateLimit `json:"rateLimit"`
	} `json:"data"`
	Errors []Error `json:"errors"`
}

type OwnerModel struct {
	*Model
}

func (o *OwnerModel) CreateIndexes() {
	database.CreateIndexes(o.Name(), []string{
		"ranks.tags",
	})
}

func (o *OwnerModel) List(tags []string, updatedAt time.Time, page int) (res []bson.M) {
	ctx := context.Background()
	limit := 10
	pipeline := mongo.Pipeline{
		bson.D{
			{"$unwind", "$ranks"},
		},
		bson.D{
			{"$match", bson.D{
				{"$and", []bson.D{{
					{"ranks.tags", tags},
					{"ranks.updated_at", updatedAt},
				}}},
			}},
		},
		bson.D{
			{"$project", bson.D{
				{"_id", "$_id"},
				{"avatar_url", "$avatar_url"},
				{"location", "$location"},
				{"name", "$name"},
				{"rank", "$ranks"},
			}},
		},
		bson.D{
			{"$sort", bson.D{
				{"rank.rank", 1},
			}},
		},
		bson.D{
			{"$skip", (page - 1) * limit},
		},
		bson.D{
			{"$limit", limit},
		},
	}
	if err := database.Aggregate(ctx, o.Name(), pipeline).All(ctx, &res); err != nil {
		log.Fatalln(err.Error())
	}
	return
}

func (o *OwnerModel) Find(id string) *mongo.SingleResult {
	projection := bson.D{
		{"gists", 0},
		{"repositories", 0},
	}
	opts := options.FindOne().SetProjection(projection)
	return database.FindOne(o.Name(), bson.D{{"_id", id}}, opts)
}

func (o *OwnerModel) Store(owners []Owner) *mongo.BulkWriteResult {
	if len(owners) == 0 {
		return nil
	}
	var models []mongo.WriteModel
	for _, owner := range owners {
		owner.TagType()
		owner.TagLocations()
		filter := bson.D{{"_id", owner.ID()}}
		update := bson.D{{"$set", owner}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}
	return database.BulkWrite(o.Name(), models)
}

func (o *OwnerModel) UpdateGists(owner Owner, gists []Gist) {
	filter := bson.D{{"_id", owner.ID()}}
	update := bson.D{{"$set", bson.D{{"gists", gists}}}}
	database.UpdateOne(o.Name(), filter, update)
}

func (o *OwnerModel) UpdateRepositories(owner Owner, repositories []Repository) {
	filter := bson.D{{"_id", owner.ID()}}
	update := bson.D{{"$set", bson.D{{"repositories", repositories}}}}
	database.UpdateOne(o.Name(), filter, update)
}

func NewOwnerModel() *OwnerModel {
	return &OwnerModel{
		&Model{
			name: "owners",
		},
	}
}
