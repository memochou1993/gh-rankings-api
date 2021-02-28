package request

import (
	"github.com/memochou1993/gh-rankings/util"
	"net/http"
	"strconv"
)

type Organization struct {
	Q     string `json:"q" validate:"omitempty"`
	Page  int64  `json:"page" validate:"omitempty,numeric"`
	Limit int64  `json:"limit" validate:"omitempty,numeric"`
}

func (o *Organization) String() string {
	return util.ParseStruct(o, ",")
}

func NewOrganizationRequest(r *http.Request) (req *Organization, err error) {
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 10
	}
	req = &Organization{
		Q:     sanitize(r.URL.Query().Get("q")),
		Page:  page,
		Limit: limit,
	}
	err = validate.Struct(req)
	return req, err
}
