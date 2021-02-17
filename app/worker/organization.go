package worker

import (
	"context"
	"errors"
	"fmt"
	"github.com/memochou1993/gh-rankings/app"
	"github.com/memochou1993/gh-rankings/app/model"
	"github.com/memochou1993/gh-rankings/app/pipeline"
	"github.com/memochou1993/gh-rankings/app/query"
	"github.com/memochou1993/gh-rankings/app/response"
	"github.com/memochou1993/gh-rankings/logger"
	"github.com/memochou1993/gh-rankings/util"
	"os"
	"strconv"
	"sync"
	"time"
)

type Organization struct {
	*Worker
	From              time.Time
	To                time.Time
	OrganizationModel *model.OrganizationModel
	RankModel         *model.RankModel
	SearchQuery       *query.Query
	RepositoryQuery   *query.Query
}

func (o *Organization) Init() {
	o.Worker.load(TimestampOrganization)
}

func (o *Organization) Collect() error {
	logger.Info("Collecting organizations...")
	o.From = time.Date(2007, time.October, 1, 0, 0, 0, 0, time.UTC)
	o.To = time.Now()

	if o.Worker.Timestamp.IsZero() {
		last := model.Organization{}
		if o.OrganizationModel.Model.Last(&last); last.ID() != "" {
			o.From = last.CreatedAt.AddDate(0, 0, -7).Truncate(24 * time.Hour)
		}
	}

	return o.Travel()
}

func (o *Organization) Travel() error {
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

func (o *Organization) Fetch(organizations *[]model.Organization) error {
	res := response.Organization{}
	if err := o.query(*o.SearchQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Search.Edges {
		*organizations = append(*organizations, edge.Node)
	}
	res.Data.RateLimit.Throttle(collecting)
	if !res.Data.Search.PageInfo.HasNextPage {
		o.SearchQuery.SearchArguments.After = ""
		return nil
	}
	o.SearchQuery.SearchArguments.After = strconv.Quote(res.Data.Search.PageInfo.EndCursor)

	return o.Fetch(organizations)
}

func (o *Organization) Update(organization model.Organization) error {
	o.RepositoryQuery.Type = model.TypeOrganization
	o.RepositoryQuery.OwnerArguments.Login = strconv.Quote(organization.ID())
	if err := o.UpdateRepositories(organization); err != nil {
		return err
	}

	return nil
}

func (o *Organization) UpdateRepositories(organization model.Organization) error {
	var repositories []model.Repository
	if err := o.FetchRepositories(&repositories); err != nil {
		return err
	}
	o.OrganizationModel.UpdateRepositories(organization, repositories)
	logger.Success(fmt.Sprintf("Updated %d %s repositories!", len(repositories), model.TypeOrganization))
	return nil
}

func (o *Organization) FetchRepositories(repositories *[]model.Repository) error {
	res := response.Organization{}
	if err := o.query(*o.RepositoryQuery, &res); err != nil {
		return err
	}
	for _, edge := range res.Data.Organization.Repositories.Edges {
		*repositories = append(*repositories, edge.Node)
	}
	res.Data.RateLimit.Throttle(collecting)
	if !res.Data.Organization.Repositories.PageInfo.HasNextPage {
		o.RepositoryQuery.RepositoriesArguments.After = ""
		return nil
	}
	o.RepositoryQuery.RepositoriesArguments.After = strconv.Quote(res.Data.Organization.Repositories.PageInfo.EndCursor)

	return o.FetchRepositories(repositories)
}

func (o *Organization) Rank() {
	logger.Info("Executing organization rank pipelines...")
	pipelines := o.buildRankPipelines()

	ch := make(chan struct{}, 2)
	wg := sync.WaitGroup{}
	wg.Add(len(pipelines))

	now := time.Now()
	for i, p := range pipelines {
		ch <- struct{}{}
		go func(p *pipeline.Pipeline) {
			defer wg.Done()
			o.RankModel.Store(o.OrganizationModel, *p, now)
			<-ch
		}(p)
		if (i+1)%10 == 0 || (i+1) == len(pipelines) {
			logger.Success(fmt.Sprintf("Executed %d of %d organization rank pipelines!", i+1, len(pipelines)))
		}
	}
	wg.Wait()
	o.Worker.save(TimestampOrganization, now)

	o.RankModel.Delete(now, model.TypeOrganization)
}

func (o *Organization) query(q query.Query, res *response.Organization) (err error) {
	if err = app.Fetch(context.Background(), fmt.Sprint(q), res); err != nil {
		if !os.IsTimeout(err) {
			return err
		}
	}
	if res.Message != "" {
		err = errors.New(res.Message)
		res.Message = ""
	}
	for _, err = range res.Errors {
		return err
	}
	if err != nil {
		logger.Error(err.Error())
		logger.Warning("Retrying...")
		time.Sleep(10 * time.Second)
		return o.query(q, res)
	}
	return
}

func (o *Organization) buildSearchQuery() string {
	from := o.From.Format(time.RFC3339)
	to := o.From.AddDate(0, 0, 7).Format(time.RFC3339)
	q := query.SearchQuery{
		Created: fmt.Sprintf("%s..%s", from, to),
		Repos:   "25..1000",
		Sort:    "joined-asc",
		Type:    model.TypeOrganization,
	}
	return strconv.Quote(util.ParseStruct(q, " "))
}

func (o *Organization) buildRankPipelines() (pipelines []*pipeline.Pipeline) {
	rankType := model.TypeOrganization
	fields := []string{
		"repositories.forks",
		"repositories.stargazers",
		"repositories.watchers",
	}
	for _, field := range fields {
		pipelines = append(pipelines, pipeline.RankByField(rankType, field))
		pipelines = append(pipelines, pipeline.RankByLocation(rankType, field)...)
		pipelines = append(pipelines, pipeline.RankOwnerRepositoryByLanguage(rankType, field)...)
	}
	return
}

func NewOrganizationWorker() *Organization {
	return &Organization{
		Worker:            &Worker{},
		OrganizationModel: model.NewOrganizationModel(),
		RankModel:         model.NewRankModel(),
		SearchQuery:       query.Owners(),
		RepositoryQuery:   query.OwnerRepositories(),
	}
}
