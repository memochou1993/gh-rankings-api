package request

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Title     string
	Tags      []string
	Timestamp time.Time
	Page      int64
	Limit     int64
}

func Parse(r *http.Request) *Request {
	title := r.URL.Query().Get("title")
	tags := strings.Split(r.URL.Query().Get("tags"), ",")
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil || page < 0 {
		page = 1
	}
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil || limit < 0 {
		limit = 10
	}
	if title == "" && limit > 100 {
		limit = 1000
	}
	return &Request{
		Title: title,
		Tags:  tags,
		Page:  page,
		Limit: limit,
	}
}
