package organization

import (
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/query"
	"github.com/memochou1993/gh-rankings/app/worker"
	"github.com/memochou1993/gh-rankings/database"
	"github.com/memochou1993/gh-rankings/test"
	"github.com/memochou1993/gh-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
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
	util.LoadEnv()
	database.Connect()
}

func TestFetch(t *testing.T) {
	o := worker.NewOrganizationWorker()

	o.SearchQuery = query.Owners()
	o.SearchQuery.SearchArguments.Query = strconv.Quote("created:2020-01-01..2020-01-01 repos:50..1000 sort:joined-asc")

	var organizations []model.Organization
	if err := o.Fetch(&organizations); err != nil {
		t.Error(err.Error())
	}
	if len(organizations) == 0 {
		t.Fail()
	}

	test.DropCollection(o.OrganizationModel)
}

func TestStore(t *testing.T) {
	o := worker.NewOrganizationWorker()

	organization := model.Organization{Login: "github"}
	organizations := []model.Organization{organization}

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

	o.RepositoryQuery = query.OwnerRepositories()
	o.RepositoryQuery.Type = model.TypeOrganization
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
