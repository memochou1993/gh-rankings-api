package owner

import (
	"github.com/memochou1993/github-rankings/app/handler"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/test"
	"github.com/memochou1993/github-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"strconv"
	"testing"
)

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	test.ChangeDirectory()
	util.LoadEnv()
	database.Init()
	logger.Init()
}

func TestFetchOwners(t *testing.T) {
	o := handler.NewOwnerHandler()

	q := model.Query{
		Schema: model.ReadQuery("search_owners"),
		SearchArguments: model.SearchArguments{
			First: 100,
			Query: strconv.Quote("created:2020-01-01..2020-01-01 followers:>=1 repos:>=10 sort:joined-asc"),
			Type:  "USER",
		},
	}

	var owners []model.Owner
	if err := o.FetchOwners(&q, &owners); err != nil {
		t.Error(err.Error())
	}
	if len(owners) == 0 {
		t.Fail()
	}

	test.DropCollection(o.OwnerModel)
}

func TestStoreUsers(t *testing.T) {
	o := handler.NewOwnerHandler()

	owner := model.Owner{Login: "memochou1993", Followers: &model.Directory{TotalCount: 1}}
	owners := []model.Owner{owner}
	o.OwnerModel.Store(owners)
	res := database.FindOne(o.OwnerModel.Name(), bson.D{{"_id", owner.ID()}})
	if res.Err() == mongo.ErrNoDocuments {
		t.Fail()
	}

	owner = model.Owner{}
	if err := res.Decode(&owner); err != nil {
		t.Fatal()
	}
	if owner.Type != model.TypeUser {
		t.Fail()
	}

	test.DropCollection(o.OwnerModel)
}

func TestStoreOrganizations(t *testing.T) {
	o := handler.NewOwnerHandler()

	organization := model.Owner{Login: "github"}
	organizations := []model.Owner{organization}
	o.OwnerModel.Store(organizations)
	res := database.FindOne(o.OwnerModel.Name(), bson.D{{"_id", organization.ID()}})
	if res.Err() == mongo.ErrNoDocuments {
		t.Fail()
	}

	organization = model.Owner{}
	if err := res.Decode(&organization); err != nil {
		t.Fatal()
	}
	if organization.Type != model.TypeOrganization {
		t.Fail()
	}

	test.DropCollection(o.OwnerModel)
}

func TestFetchUserRepositories(t *testing.T) {
	o := handler.NewOwnerHandler()

	q := model.Query{
		Schema: model.ReadQuery("owner_repositories"),
		Field:  model.TypeUser,
		OwnerArguments: model.OwnerArguments{
			Login: strconv.Quote("memochou1993"),
		},
		RepositoriesArguments: model.RepositoriesArguments{
			First:             100,
			OrderBy:           "{field:CREATED_AT,direction:ASC}",
			OwnerAffiliations: "OWNER",
		},
	}

	var repositories []model.Repository
	if err := o.FetchRepositories(&q, &repositories); err != nil {
		t.Error(err.Error())
	}
	if len(repositories) == 0 {
		t.Fail()
	}

	test.DropCollection(o.OwnerModel)
}

func TestFetchOrganizationRepositories(t *testing.T) {
	o := handler.NewOwnerHandler()

	q := model.Query{
		Schema: model.ReadQuery("owner_repositories"),
		Field:  model.TypeOrganization,
		OwnerArguments: model.OwnerArguments{
			Login: strconv.Quote("facebook"),
		},
		RepositoriesArguments: model.RepositoriesArguments{
			First:             100,
			OrderBy:           "{field:CREATED_AT,direction:ASC}",
			OwnerAffiliations: "OWNER",
		},
	}

	var repositories []model.Repository
	if err := o.FetchRepositories(&q, &repositories); err != nil {
		t.Error(err.Error())
	}
	if len(repositories) == 0 {
		t.Fail()
	}

	test.DropCollection(o.OwnerModel)
}

func tearDown() {
	test.DropDatabase()
}
