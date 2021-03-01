package model

import (
	"github.com/memochou1993/gh-rankings/app/handler/request"
	"github.com/memochou1993/gh-rankings/app/pipeline"
	"github.com/memochou1993/gh-rankings/app/resource"
	"github.com/memochou1993/gh-rankings/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"log"
	"time"
)

type Organization struct {
	ImageUrl       string       `json:"imageUrl" bson:"image_url"`
	CreatedAt      *time.Time   `json:"createdAt" bson:"created_at"`
	Location       string       `json:"location" bson:"location"`
	Login          string       `json:"login" bson:"_id"`
	Name           string       `json:"name" bson:"name"`
	Repositories   []Repository `json:"repositories,omitempty" bson:"repositories,omitempty"`
	ParsedLocation string       `json:"parsedLocation" bson:"parsed_location"`
	ParsedCity     string       `json:"parsedCity" bson:"parsed_city"`
}

func (o *Organization) ID() string {
	return o.Login
}

func (o *Organization) parseLocation() {
	o.ParsedLocation, o.ParsedCity = resource.Locate(o.Location)
}

type OrganizationModel struct {
	*Model
}

func (o *OrganizationModel) List(req *request.Organization) (organizations []Organization) {
	ctx := context.Background()

	p := pipeline.ListOrganizations(req)
	if req.Q != "" {
		p = pipeline.SearchOrganizations(req)
	}

	cursor := database.Aggregate(ctx, o.Model.Name(), p)
	organizations = make([]Organization, req.Limit)
	if err := cursor.All(ctx, &organizations); err != nil {
		log.Fatal(err.Error())
	}

	return
}

func (o *OrganizationModel) FindByID(id string) (organization Organization) {
	o.Model.FindByID(id, &organization)
	return
}

func (o *OrganizationModel) Store(organizations []Organization) *mongo.BulkWriteResult {
	if len(organizations) == 0 {
		return nil
	}
	var models []mongo.WriteModel
	for _, organization := range organizations {
		organization.parseLocation()
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
		Model: &Model{
			name: "organizations",
		},
	}
}
