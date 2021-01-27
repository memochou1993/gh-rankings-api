package model

import (
	"fmt"
	"github.com/memochou1993/github-rankings/app/resource"
	"github.com/memochou1993/github-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Owner struct {
	AvatarURL    string       `json:"avatarUrl,omitempty" bson:"avatar_url,omitempty"`
	CreatedAt    *time.Time   `json:"createdAt,omitempty" bson:"created_at,omitempty"`
	Followers    *Directory   `json:"followers,omitempty" bson:"followers,omitempty"`
	Location     string       `json:"location,omitempty" bson:"location,omitempty"`
	Login        string       `json:"login" bson:"_id"`
	Name         string       `json:"name,omitempty" bson:"name,omitempty"`
	Gists        []Gist       `json:"gists,omitempty" bson:"gists,omitempty"`
	Repositories []Repository `json:"repositories,omitempty" bson:"repositories,omitempty"`
	Tags         []string     `json:"tags,omitempty" bson:"tags,omitempty"`
}

func (o *Owner) ID() string {
	return o.Login
}

func (o *Owner) IsFollowersEmpty() bool {
	return o.Followers == nil
}

func (o *Owner) IsUser() bool {
	return !o.IsFollowersEmpty()
}

func (o *Owner) IsOrganization() bool {
	return o.IsFollowersEmpty()
}

func (o *Owner) TagType() {
	o.Tags = append(o.Tags, fmt.Sprintf("type:%s", o.Type()))
}

func (o *Owner) TagLocations() {
	for _, location := range resource.Locate(o.Location) {
		o.Tags = append(o.Tags, fmt.Sprintf("location:%s", location))
	}
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
			AvatarURL string     `json:"avatarUrl"`
			CreatedAt *time.Time `json:"createdAt"`
			Followers *Directory `json:"followers"`
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
