package user

import (
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/worker"
	"github.com/memochou1993/gh-rankings/database"
	"github.com/memochou1993/gh-rankings/logger"
	"github.com/memochou1993/gh-rankings/test"
	"github.com/memochou1993/gh-rankings/util"
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

func TestFetchOrganizations(t *testing.T) {
	o := worker.NewOrganizationWorker()

	o.SearchQuery = model.NewOwnerQuery()
	o.SearchQuery.SearchArguments.Query = strconv.Quote("created:2020-01-01..2020-01-01 repos:>=50 sort:joined-asc")

	organizations := map[string]model.Organization{}
	if err := o.FetchOrganizations(organizations); err != nil {
		t.Error(err.Error())
	}
	if len(organizations) == 0 {
		t.Fail()
	}

	test.DropCollection(o.OrganizationModel)
}

func TestStoreOrganizations(t *testing.T) {
	o := worker.NewOrganizationWorker()

	organization := model.Organization{Login: "github"}
	organizations := map[string]model.Organization{}
	organizations[organization.Login] = organization

	o.OrganizationModel.Store(organizations)
	res := database.FindOne(o.OrganizationModel.Name(), bson.D{{"_id", organization.ID()}})
	if res.Err() == mongo.ErrNoDocuments {
		t.Fail()
	}

	organization = model.Organization{}
	if err := res.Decode(&organization); err != nil {
		t.Fatal(err.Error())
	}

	test.DropCollection(o.OrganizationModel)
}

func TestFetchRepositories(t *testing.T) {
	o := worker.NewUserWorker()

	o.RepositoryQuery = model.NewOwnerRepositoryQuery()
	o.RepositoryQuery.Field = model.TypeOrganization
	o.RepositoryQuery.OwnerArguments.Login = strconv.Quote("facebook")

	var repositories []model.Repository
	if err := o.FetchRepositories(&repositories); err != nil {
		t.Error(err.Error())
	}
	if len(repositories) == 0 {
		t.Fail()
	}

	test.DropCollection(o.UserModel)
}

func tearDown() {
	test.DropDatabase()
}
