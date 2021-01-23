package request

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Name      string
	Tags      []string
	Timestamp time.Time
	Page      int64
	Limit     int64
}

func Parse(r *http.Request) *Request {
	name := r.URL.Query().Get("name")
	tags := strings.Split(r.URL.Query().Get("tags"), ",")
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil || page < 0 {
		page = 1
	}
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil || limit < 0 {
		limit = 10
	}
	if name == "" && limit > 1000 {
		limit = 1000
	}
	return &Request{
		Name:  name,
		Tags:  tags,
		Page:  page,
		Limit: limit,
	}
}
