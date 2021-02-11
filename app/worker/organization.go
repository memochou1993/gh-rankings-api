package worker

import (
	"context"
	"errors"
	"fmt"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/resource"
	"github.com/memochou1993/gh-rankings/app/response"
	"github.com/memochou1993/gh-rankings/logger"
	"github.com/memochou1993/gh-rankings/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"strconv"
	"sync"
	"time"
)

type organizationWorker struct {
	*Worker
	From              time.Time
	To                time.Time
	OrganizationModel *model.OrganizationModel
	SearchQuery       *model.Query
	RepositoryQuery   *model.Query
}

func (o *organizationWorker) Collect() error {
	logger.Info("Collecting organizations...")
	o.From = time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	o.To = time.Now()

	if o.Worker.Timestamp.IsZero() {
		if organization := o.OrganizationModel.FindLast(); organization.ID() != "" {
			o.From = organization.CreatedAt.AddDate(0, 0, -7).Truncate(24 * time.Hour)
		}
	}

	return o.Travel()
}

func (o *organizationWorker) Travel() error {
	if o.From.After(o.To) {
		return nil
	}

	var organizations []model.Organization
	o.SearchQuery.SearchArguments.Query = o.buildSearchQuery()
	logger.Debug(fmt.Sprintf("Organization Query: %s", o.SearchQuery.SearchArguments.Query))
	if err := o.Fetch(&organizations); err != nil {
		return err
	}

	if res := o.OrganizationModel.Store(organizations); res != nil {
		if res.ModifiedCount > 0 {
			logger.Success(fmt.Sprintf("Updated %d organizations!", res.ModifiedCount))
		}
		if res.UpsertedCount > 0 {
			logger.Success(fmt.Sprintf("Inserted %d organizations!", res.UpsertedCount))
		}
	}
	for _, organization := range organizations {
		if err := o.Update(organization); err != nil {
			return err
		}
	}
	o.From = o.From.AddDate(0, 0, 7)

	return o.Travel()
}

func (o *organizationWorker) Fetch(organizations *[]model.Organization) error {
	res := response.Organization{}
	if err := o.query(*o.SearchQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		*organizations = append(*organizations, edge.Node)
	}
	res.Data.RateLimit.Break(collecting)
	if !res.Data.Search.PageInfo.HasNextPage {
		o.SearchQuery.SearchArguments.After = ""
		return nil
	}
	o.SearchQuery.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return o.Fetch(organizations)
}

func (o *organizationWorker) Update(organization model.Organization) error {
	o.RepositoryQuery.Field = model.TypeOrganization
	o.RepositoryQuery.OwnerArguments.Login = strconv.Quote(organization.ID())
	if err := o.UpdateRepositories(organization); err != nil {
		return err
	}

	return nil
}

func (o *organizationWorker) UpdateRepositories(organization model.Organization) error {
	var repositories []model.Repository
	if err := o.FetchRepositories(&repositories); err != nil {
		return err
	}
	o.OrganizationModel.UpdateRepositories(organization, repositories)
	logger.Success(fmt.Sprintf("Updated %d %s repositories!", len(repositories), model.TypeOrganization))
	return nil
}

func (o *organizationWorker) FetchRepositories(repositories *[]model.Repository) error {
	res := response.Organization{}
	if err := o.query(*o.RepositoryQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Organization.Repositories.Edges {
		*repositories = append(*repositories, edge.Node)
	}
	res.Data.RateLimit.Break(collecting)
	if !res.Data.Organization.Repositories.PageInfo.HasNextPage {
		o.RepositoryQuery.RepositoriesArguments.After = ""
		return nil
	}
	o.RepositoryQuery.RepositoriesArguments.After = strconv.Quote(res.Data.Organization.Repositories.PageInfo.EndCursor)

	return o.FetchRepositories(repositories)
}

func (o *organizationWorker) Rank() {
	logger.Info("Executing organization rank pipelines...")
	var pipelines []*model.Pipeline
	pipelines = append(pipelines, o.buildRankPipelines()...)

	ch := make(chan struct{}, 2)
	wg := sync.WaitGroup{}
	wg.Add(len(pipelines))

	now := time.Now()
	for i, p := range pipelines {
		ch <- struct{}{}
		go func(p *model.Pipeline) {
			defer wg.Done()
			RankModel.Store(o.OrganizationModel, *p, now)
			<-ch
		}(p)
		if (i+1)%10 == 0 || (i+1) == len(pipelines) {
			logger.Success(fmt.Sprintf("Executed %d of %d organization rank pipelines!", i+1, len(pipelines)))
		}
	}
	wg.Wait()
	o.Worker.seal(timestampOrganizationRanks, now)

	tags := []string{fmt.Sprintf("type:%s", model.TypeOrganization)}
	RankModel.Delete(now, tags...)
}

func (o *organizationWorker) query(q model.Query, res *response.Organization) (err error) {
	if err = app.Fetch(context.Background(), fmt.Sprint(q), res); err != nil {
		if !os.IsTimeout(err) {
			return err
		}
	}
	if res.Message != "" {
		err = errors.New(res.Message)
	}
	for _, err = range res.Errors {
		break
	}
	if err != nil {
		logger.Error(err.Error())
		logger.Warning("Retrying...")
		time.Sleep(10 * time.Second)
		return o.query(q, res)
	}
	return
}

func (o *organizationWorker) buildSearchQuery() string {
	from := o.From.Format(time.RFC3339)
	to := o.From.AddDate(0, 0, 7).Format(time.RFC3339)
	q := model.SearchQuery{
		Created: fmt.Sprintf("%s..%s", from, to),
		Repos:   ">=25",
		Sort:    "joined-asc",
		Type:    model.TypeOrganization,
	}
	return strconv.Quote(util.ParseStruct(q, " "))
}

func (o *organizationWorker) buildRankPipelines() (pipelines []*model.Pipeline) {
	tag := fmt.Sprintf("type:%s", model.TypeOrganization)
	fields := []string{
		"repositories.forks",
		"repositories.stargazers",
		"repositories.watchers",
	}
	for _, field := range fields {
		pipelines = append(pipelines, o.buildRankPipeline(field, tag))
		pipelines = append(pipelines, o.buildRankPipelinesByLocation(field, tag)...)
	}
	pipelines = append(pipelines, o.buildRepositoryRankPipelinesByLanguage("forks", tag)...)
	pipelines = append(pipelines, o.buildRepositoryRankPipelinesByLanguage("stargazers", tag)...)
	pipelines = append(pipelines, o.buildRepositoryRankPipelinesByLanguage("watchers", tag)...)
	return
}

func (o *organizationWorker) buildRankPipeline(field string, tags ...string) *model.Pipeline {
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

func (o *organizationWorker) buildRankPipelinesByLocation(field string, tags ...string) (pipelines []*model.Pipeline) {
	for _, location := range resource.Locations {
		pipelines = append(pipelines, o.buildRankPipeline(field, append(tags, fmt.Sprintf("location:%s", location.Name))...))
		for _, city := range location.Cities {
			pipelines = append(pipelines, o.buildRankPipeline(field, append(tags, fmt.Sprintf("location:%s, %s", city.Name, location.Name))...))
		}
	}
	return
}

func (o *organizationWorker) buildRepositoryRankPipelinesByLanguage(field string, tags ...string) (pipelines []*model.Pipeline) {
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

func NewOrganizationWorker() *organizationWorker {
	return &organizationWorker{
		Worker:            NewWorker(),
		OrganizationModel: model.NewOrganizationModel(),
		SearchQuery:       model.NewOwnerQuery(),
		RepositoryQuery:   model.NewOwnerRepositoryQuery(),
	}
}
