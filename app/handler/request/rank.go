package request

import (
	"github.com/memochou1993/gh-rankings/util"
	"net/http"
	"strconv"
	"time"
)

type Rank struct {
	Name       string `json:"name" validate:"omitempty"`
	Field      string `json:"field" validate:"omitempty"`
	Type       string `json:"type" validate:"omitempty,alpha"`
	Language   string `json:"language" validate:"omitempty"`
	Location   string `json:"location" validate:"omitempty"`
	Page       int64  `json:"page" validate:"omitempty,numeric"`
	Limit      int64  `json:"limit" validate:"omitempty,numeric"`
	Timestamps []time.Time
}

func (r *Rank) String() string {
	return util.ParseStruct(r, ",")
}

func NewRankRequest(r *http.Request) (req *Rank, err error) {
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 10
	}
	req = &Rank{
		Name:     sanitize(r.URL.Query().Get("name")),
		Type:     sanitize(r.URL.Query().Get("type")),
		Field:    sanitize(r.URL.Query().Get("field")),
		Language: sanitize(r.URL.Query().Get("language")),
		Location: sanitize(r.URL.Query().Get("location")),
		Page:     page,
		Limit:    limit,
	}
	err = validate.Struct(req)
	return req, err
}
