package owner

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

func TestFetchUsers(t *testing.T) {
	o := worker.NewOwnerWorker()

	o.UserQuery = model.NewOwnerQuery()
	o.UserQuery.SearchArguments.Query = strconv.Quote("created:2020-01-01..2020-01-01 followers:>=50 repos:>=5 sort:joined-asc")

	owners := map[string]model.Owner{}
	if err := o.FetchUsers(owners); err != nil {
		t.Error(err.Error())
	}

	users := map[string]model.Owner{}
	for _, owner := range owners {
		if owner.IsUser() {
			users[owner.Login] = owner
		}
	}
	if len(users) == 0 {
		t.Fail()
	}

	test.DropCollection(o.OwnerModel)
}

func TestFetchOrganizations(t *testing.T) {
	o := worker.NewOwnerWorker()

	o.OrganizationQuery = model.NewOwnerQuery()
	o.OrganizationQuery.SearchArguments.Query = strconv.Quote("created:2020-01-01..2020-01-01 repos:>=50 sort:joined-asc")

	owners := map[string]model.Owner{}
	if err := o.FetchOrganizations(owners); err != nil {
		t.Error(err.Error())
	}

	organizations := map[string]model.Owner{}
	for _, owner := range owners {
		if owner.IsOrganization() {
			organizations[owner.Login] = owner
		}
	}
	if len(organizations) == 0 {
		t.Fail()
	}

	test.DropCollection(o.OwnerModel)
}

func TestStoreUsers(t *testing.T) {
	o := worker.NewOwnerWorker()

	owner := model.Owner{Login: "memochou1993", Followers: &model.Directory{TotalCount: 1}}
	owners := map[string]model.Owner{}
	owners["memochou1993"] = owner

	o.OwnerModel.Store(owners)
	res := database.FindOne(o.OwnerModel.Name(), bson.D{{"_id", owner.ID()}})
	if res.Err() == mongo.ErrNoDocuments {
		t.Fail()
	}

	owner = model.Owner{}
	if err := res.Decode(&owner); err != nil {
		t.Fatal(err.Error())
	}
	if !owner.IsUser() {
		t.Fail()
	}

	test.DropCollection(o.OwnerModel)
}

func TestStoreOrganizations(t *testing.T) {
	o := worker.NewOwnerWorker()

	owner := model.Owner{Login: "github"}
	owners := map[string]model.Owner{}
	owners[owner.Login] = owner

	o.OwnerModel.Store(owners)
	res := database.FindOne(o.OwnerModel.Name(), bson.D{{"_id", owner.ID()}})
	if res.Err() == mongo.ErrNoDocuments {
		t.Fail()
	}

	owner = model.Owner{}
	if err := res.Decode(&owner); err != nil {
		t.Fatal(err.Error())
	}
	if !owner.IsOrganization() {
		t.Fail()
	}

	test.DropCollection(o.OwnerModel)
}

func TestFetchGists(t *testing.T) {
	o := worker.NewOwnerWorker()

	o.GistQuery = model.NewOwnerGistQuery()
	o.GistQuery.OwnerArguments.Login = strconv.Quote("memochou1993")

	var gists []model.Gist
	if err := o.FetchGists(&gists); err != nil {
		t.Error(err.Error())
	}
	if len(gists) == 0 {
		t.Fail()
	}

	test.DropCollection(o.OwnerModel)
}

func TestFetchRepositories(t *testing.T) {
	o := worker.NewOwnerWorker()

	o.RepositoryQuery = model.NewOwnerRepositoryQuery()
	o.RepositoryQuery.Field = model.TypeUser
	o.RepositoryQuery.OwnerArguments.Login = strconv.Quote("memochou1993")

	var repositories []model.Repository
	if err := o.FetchRepositories(&repositories); err != nil {
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
