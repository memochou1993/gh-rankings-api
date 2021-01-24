package request

import (
	"net/http"
	"strconv"
	"strings"
)

type Request struct {
	Name  string
	Tags  []string
	Page  int64
	Limit int64
}

func (r *Request) HasTag(tag string) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (r *Request) IsNameEmpty() bool {
	return r.Name == ""
}

func (r *Request) IsTagsEmpty() bool {
	return len(r.Tags) == 0
}

func (r *Request) IsEmpty() bool {
	return r.IsNameEmpty() && r.IsTagsEmpty()
}

func Parse(r *http.Request) *Request {
	name := r.URL.Query().Get("name")
	var tags []string
	for _, tag := range strings.Split(r.URL.Query().Get("tags"), ",") {
		if tag != "" {
			tags = append(tags, tag)
		}
	}
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
