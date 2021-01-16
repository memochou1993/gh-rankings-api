package request

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type RepositoryRequest struct {
	NameWithOwner string
	Tags          []string
	CreatedAt     *time.Time
	Page          int64
	Limit         int64
}

func NewRepositoryRequest(r *http.Request) *RepositoryRequest {
	nameWithOwner := r.URL.Query().Get("nameWithOwner")
	tags := strings.Split(r.URL.Query().Get("tags"), ",")
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil || page < 0 {
		page = 1
	}
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil || limit < 0 {
		limit = 10
	}
	if nameWithOwner == "" && limit > 100 {
		limit = 1000
	}
	return &RepositoryRequest{
		NameWithOwner: nameWithOwner,
		Tags:          tags,
		Page:          page,
		Limit:         limit,
	}
}
