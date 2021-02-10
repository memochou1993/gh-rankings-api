package model

import (
	"fmt"
	"github.com/memochou1993/gh-rankings/app/resource"
	"github.com/memochou1993/gh-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Organization struct {
	AvatarURL    string       `json:"avatarUrl,omitempty" bson:"avatar_url,omitempty"`
	CreatedAt    *time.Time   `json:"createdAt,omitempty" bson:"created_at,omitempty"`
	Location     string       `json:"location,omitempty" bson:"location,omitempty"`
	Login        string       `json:"login" bson:"_id"`
	Name         string       `json:"name,omitempty" bson:"name,omitempty"`
	Repositories []Repository `json:"repositories,omitempty" bson:"repositories,omitempty"`
	Tags         []string     `json:"tags,omitempty" bson:"tags,omitempty"`
}

func (o *Organization) ID() string {
	return o.Login
}

// TODO: should remove
func (o *Organization) Type() string {
	return TypeOrganization
}

// TODO: should remove
func (o *Organization) TagType() {
	o.Tags = append(o.Tags, fmt.Sprintf("type:%s", TypeOrganization))
}

func (o *Organization) TagLocations() {
	for _, location := range resource.Locate(o.Location) {
		o.Tags = append(o.Tags, fmt.Sprintf("location:%s", location))
	}
}

type OrganizationResponse struct {
	Data struct {
		Search struct {
			Edges []struct {
				Cursor string       `json:"cursor"`
				Node   Organization `json:"node"`
			} `json:"edges"`
			PageInfo `json:"pageInfo"`
		} `json:"search"`
		Organization struct {
			AvatarURL    string     `json:"avatarUrl"`
			CreatedAt    *time.Time `json:"createdAt"`
			Location     string     `json:"location"`
			Login        string     `json:"login"`
			Name         string     `json:"name"`
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

type OrganizationModel struct {
	*Model
}

func (o *OrganizationModel) Store(organizations map[string]Organization) *mongo.BulkWriteResult {
	if len(organizations) == 0 {
		return nil
	}
	var models []mongo.WriteModel
	for _, organization := range organizations {
		organization.TagType()
		organization.TagLocations()
		filter := bson.D{{"_id", organization.ID()}}
		update := bson.D{{"$set", organization}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}
	return database.BulkWrite(o.Name(), models)
}

func (o *OrganizationModel) UpdateRepositories(organization Organization, repositories []Repository) {
	filter := bson.D{{"_id", organization.ID()}}
	update := bson.D{{"$set", bson.D{{"repositories", repositories}}}}
	database.UpdateOne(o.Name(), filter, update)
}

func NewOrganizationModel() *OrganizationModel {
	return &OrganizationModel{
		&Model{
			name: "organizations",
		},
	}
}
