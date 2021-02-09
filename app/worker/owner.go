package worker

import (
	"context"
	"fmt"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/resource"
	"github.com/memochou1993/gh-rankings/logger"
	"github.com/memochou1993/gh-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	OwnerWorker *ownerWorker
)

type ownerWorker struct {
	*Worker
	From              time.Time
	To                time.Time
	OwnerModel        *model.OwnerModel
	UserQuery         *model.Query
	OrganizationQuery *model.Query
	GistQuery         *model.Query
	RepositoryQuery   *model.Query
}

func (o *ownerWorker) Init() {
	o.Worker.loadTimestamp(timestampOwnerRanks)
}

func (o *ownerWorker) Collect() error {
	logger.Info("Collecting owners...")
	o.From = time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	o.To = time.Now()

	return o.Travel()
}

func (o *ownerWorker) Travel() error {
	if o.From.After(o.To) {
		return nil
	}

	owners := map[string]model.Owner{}

	o.UserQuery.SearchArguments.Query = o.buildUserSearchQuery()
	logger.Debug(fmt.Sprintf("User Query: %s", o.UserQuery.SearchArguments.Query))
	if err := o.FetchUsers(owners); err != nil {
		return err
	}

	o.OrganizationQuery.SearchArguments.Query = o.buildOrganizationSearchQuery()
	logger.Debug(fmt.Sprintf("Organization Query: %s", o.OrganizationQuery.SearchArguments.Query))
	if err := o.FetchOrganizations(owners); err != nil {
		return err
	}

	if res := o.OwnerModel.Store(owners); res != nil {
		if res.ModifiedCount > 0 {
			logger.Success(fmt.Sprintf("Updated %d owners!", res.ModifiedCount))
		}
		if res.UpsertedCount > 0 {
			logger.Success(fmt.Sprintf("Inserted %d owners!", res.UpsertedCount))
		}
	}
	for _, owner := range owners {
		if err := o.Update(owner); err != nil {
			return err
		}
	}
	o.From = o.From.AddDate(0, 0, 7)

	return o.Travel()
}

func (o *ownerWorker) FetchUsers(users map[string]model.Owner) error {
	res := model.OwnerResponse{}
	if err := o.fetch(*o.UserQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		users[edge.Node.Login] = edge.Node
	}
	res.Data.RateLimit.Check()
	if !res.Data.Search.PageInfo.HasNextPage {
		o.UserQuery.SearchArguments.After = ""
		return nil
	}
	o.UserQuery.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return o.FetchUsers(users)
}

func (o *ownerWorker) FetchOrganizations(organizations map[string]model.Owner) error {
	res := model.OwnerResponse{}
	if err := o.fetch(*o.OrganizationQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		organizations[edge.Node.Login] = edge.Node
	}
	res.Data.RateLimit.Check()
	if !res.Data.Search.PageInfo.HasNextPage {
		o.OrganizationQuery.SearchArguments.After = ""
		return nil
	}
	o.OrganizationQuery.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return o.FetchOrganizations(organizations)
}

func (o *ownerWorker) Update(owner model.Owner) error {
	o.GistQuery.OwnerArguments.Login = strconv.Quote(owner.ID())
	if err := o.UpdateGists(owner); err != nil {
		return err
	}

	o.RepositoryQuery.Field = owner.Type()
	o.RepositoryQuery.OwnerArguments.Login = strconv.Quote(owner.ID())
	if err := o.UpdateRepositories(owner); err != nil {
		return err
	}

	return nil
}

func (o *ownerWorker) UpdateGists(owner model.Owner) error {
	if !owner.IsUser() {
		return nil
	}
	var gists []model.Gist
	if err := o.FetchGists(&gists); err != nil {
		return err
	}
	o.OwnerModel.UpdateGists(owner, gists)
	logger.Success(fmt.Sprintf("Updated %d %s gists!", len(gists), owner.Type()))
	return nil
}

func (o *ownerWorker) UpdateRepositories(owner model.Owner) error {
	var repositories []model.Repository
	if err := o.FetchRepositories(&repositories); err != nil {
		return err
	}
	o.OwnerModel.UpdateRepositories(owner, repositories)
	logger.Success(fmt.Sprintf("Updated %d %s repositories!", len(repositories), owner.Type()))
	return nil
}

func (o *ownerWorker) FetchGists(gists *[]model.Gist) error {
	res := model.OwnerResponse{}
	if err := o.fetch(*o.GistQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Owner.Gists.Edges {
		*gists = append(*gists, edge.Node)
	}
	res.Data.RateLimit.Check()
	if !res.Data.Owner.Gists.PageInfo.HasNextPage {
		o.GistQuery.GistsArguments.After = ""
		return nil
	}
	o.GistQuery.GistsArguments.After = strconv.Quote(res.Data.Owner.Gists.PageInfo.EndCursor)

	return o.FetchGists(gists)
}

func (o *ownerWorker) FetchRepositories(repositories *[]model.Repository) error {
	res := model.OwnerResponse{}
	if err := o.fetch(*o.RepositoryQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Owner.Repositories.Edges {
		*repositories = append(*repositories, edge.Node)
	}
	res.Data.RateLimit.Check()
	if !res.Data.Owner.Repositories.PageInfo.HasNextPage {
		o.RepositoryQuery.RepositoriesArguments.After = ""
		return nil
	}
	o.RepositoryQuery.RepositoriesArguments.After = strconv.Quote(res.Data.Owner.Repositories.PageInfo.EndCursor)

	return o.FetchRepositories(repositories)
}

func (o *ownerWorker) Rank() {
	logger.Info("Executing owner rank pipelines...")
	var pipelines []*model.Pipeline
	pipelines = append(pipelines, o.newUserRankPipelines()...)
	pipelines = append(pipelines, o.newOrganizationRankPipelines()...)

	ch := make(chan struct{}, 2)
	wg := sync.WaitGroup{}
	wg.Add(len(pipelines))

	now := time.Now()
	for i, p := range pipelines {
		ch <- struct{}{}
		go func(p *model.Pipeline) {
			defer wg.Done()
			RankModel.Store(o.OwnerModel, *p, now)
			<-ch
		}(p)
		if (i+1)%100 == 0 || (i+1) == len(pipelines) {
			logger.Success(fmt.Sprintf("Executed %d of %d owner rank pipelines!", i+1, len(pipelines)))
		}
	}
	wg.Wait()
	o.Worker.saveTimestamp(timestampOwnerRanks, now)

	tags := []string{fmt.Sprintf("type:%s", model.TypeUser), fmt.Sprintf("type:%s", model.TypeOrganization)}
	RankModel.Delete(now, tags...)
}

func (o *ownerWorker) fetch(q model.Query, res *model.OwnerResponse) (err error) {
	if err := app.Fetch(context.Background(), fmt.Sprint(q), res); err != nil {
		if os.IsTimeout(err) {
			logger.Error("Retrying...")
			return o.fetch(q, res)
		}
		return err
	}
	for _, err := range res.Errors {
		return err
	}
	return
}

func (o *ownerWorker) buildUserSearchQuery() string {
	from := o.From.Format(time.RFC3339)
	to := o.From.AddDate(0, 0, 7).Format(time.RFC3339)
	q := model.SearchQuery{
		Created:   fmt.Sprintf("%s..%s", from, to),
		Followers: ">=100",
		Sort:      "joined-asc",
	}
	return strconv.Quote(util.ParseStruct(q, " "))
}

func (o *ownerWorker) buildOrganizationSearchQuery() string {
	from := o.From.Format(time.RFC3339)
	to := o.From.AddDate(0, 0, 7).Format(time.RFC3339)
	q := model.SearchQuery{
		Created: fmt.Sprintf("%s..%s", from, to),
		Repos:   ">=5",
		Sort:    "joined-asc",
	}
	return strconv.Quote(util.ParseStruct(q, " "))
}

func (o *ownerWorker) newUserRankPipelines() (pipelines []*model.Pipeline) {
	tag := fmt.Sprintf("type:%s", model.TypeUser)
	fields := []string{
		"followers",
		"gists.forks",
		"gists.stargazers",
		"repositories.forks",
		"repositories.stargazers",
		"repositories.watchers",
	}
	for _, field := range fields {
		pipelines = append(pipelines, o.newRankPipeline(field, tag))
		pipelines = append(pipelines, o.newRankPipelinesByLocation(field, tag)...)
	}
	pipelines = append(pipelines, o.newRepositoryRankPipelinesByLanguage("forks", tag)...)
	pipelines = append(pipelines, o.newRepositoryRankPipelinesByLanguage("stargazers", tag)...)
	pipelines = append(pipelines, o.newRepositoryRankPipelinesByLanguage("watchers", tag)...)
	return
}

func (o *ownerWorker) newOrganizationRankPipelines() (pipelines []*model.Pipeline) {
	tag := fmt.Sprintf("type:%s", model.TypeOrganization)
	fields := []string{
		"repositories.forks",
		"repositories.stargazers",
		"repositories.watchers",
	}
	for _, field := range fields {
		pipelines = append(pipelines, o.newRankPipeline(field, tag))
		pipelines = append(pipelines, o.newRankPipelinesByLocation(field, tag)...)
	}
	pipelines = append(pipelines, o.newRepositoryRankPipelinesByLanguage("forks", tag)...)
	pipelines = append(pipelines, o.newRepositoryRankPipelinesByLanguage("stargazers", tag)...)
	pipelines = append(pipelines, o.newRepositoryRankPipelinesByLanguage("watchers", tag)...)
	return
}

func (o *ownerWorker) newRankPipeline(field string, tags ...string) *model.Pipeline {
	return &model.Pipeline{
		Pipeline: &mongo.Pipeline{
			bson.D{
				{"$match", bson.D{
					{"tags", bson.D{
						{"$all", tags},
					}},
				}},
			},
			bson.D{
				{"$project", bson.D{
					{"_id", "$_id"},
					{"image_url", "$avatar_url"},
					{"total_count", bson.D{
						{"$sum", fmt.Sprintf("$%s.total_count", field)},
					}},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"total_count", -1},
				}},
			},
		},
		Tags: append(tags, fmt.Sprintf("field:%s", field)),
	}
}

func (o *ownerWorker) newRankPipelinesByLocation(field string, tags ...string) (pipelines []*model.Pipeline) {
	for _, location := range resource.Locations {
		pipelines = append(pipelines, o.newRankPipeline(field, append(tags, fmt.Sprintf("location:%s", location.Name))...))
		for _, city := range location.Cities {
			pipelines = append(pipelines, o.newRankPipeline(field, append(tags, fmt.Sprintf("location:%s, %s", city.Name, location.Name))...))
		}
	}
	return
}

func (o *ownerWorker) newRepositoryRankPipelinesByLanguage(field string, tags ...string) (pipelines []*model.Pipeline) {
	for _, language := range resource.Languages {
		pipelines = append(pipelines, &model.Pipeline{
			Pipeline: &mongo.Pipeline{
				bson.D{
					{"$match", bson.D{
						{"tags", bson.D{
							{"$all", tags},
						}},
					}},
				},
				bson.D{
					{"$unwind", "$repositories"},
				},
				bson.D{
					{"$match", bson.D{
						{"repositories.primary_language.name", language.Name},
					}},
				},
				bson.D{
					{"$group", bson.D{
						{"_id", "$_id"},
						{"image_url", bson.D{
							{"$first", "$avatar_url"},
						}},
						{"total_count", bson.D{
							{"$sum", fmt.Sprintf("$repositories.%s.total_count", field)},
						}},
					}},
				},
				bson.D{
					{"$sort", bson.D{
						{"total_count", -1},
					}},
				},
			},
			Tags: append(tags, fmt.Sprintf("field:repositories.%s", field), fmt.Sprintf("language:%s", language.Name)),
		})
	}
	return
}

func NewOwnerWorker() *ownerWorker {
	return &ownerWorker{
		Worker:            NewWorker(),
		OwnerModel:        model.NewOwnerModel(),
		UserQuery:         model.NewOwnerQuery(),
		OrganizationQuery: model.NewOwnerQuery(),
		GistQuery:         model.NewOwnerGistQuery(),
		RepositoryQuery:   model.NewOwnerRepositoryQuery(),
	}
}
