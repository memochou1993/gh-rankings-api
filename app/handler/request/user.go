package request

import (
	"net/http"
	"strconv"
)

type User struct {
	Q     string `json:"q" validate:"omitempty"`
	Page  int64  `json:"page" validate:"omitempty,numeric"`
	Limit int64  `json:"limit" validate:"omitempty,numeric"`
}

func NewUserRequest(r *http.Request) (req *User, err error) {
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 10
	}
	req = &User{
		Q:     sanitize(r.URL.Query().Get("q")),
		Page:  page,
		Limit: limit,
	}
	err = validate.Struct(req)
	return req, err
}
