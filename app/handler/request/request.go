package request

import (
	"github.com/go-playground/validator/v10"
	"net/http"
	"strconv"
	"strings"
)

var (
	validate *validator.Validate
)

type Request struct {
	Name     string `json:"name" validate:"omitempty,alphanum"`
	Field    string `json:"field" validate:"omitempty"`
	Type     string `json:"type" validate:"omitempty,alpha"`
	Language string `json:"language" validate:"omitempty"`
	Location string `json:"location" validate:"omitempty"`
	Page     int64  `json:"page" validate:"omitempty,numeric"`
	Limit    int64  `json:"limit" validate:"omitempty,numeric"`
}

func init() {
	validate = validator.New()
}

func Validate(r *http.Request) (req *Request, err error) {
	for _, f := range strings.Split(r.URL.Query().Get("field"), ".") {
		if err = validate.Var(f, "omitempty,alpha"); err != nil {
			return
		}
	}
	for _, f := range strings.Split(r.URL.Query().Get("language"), " ") {
		if err = validate.Var(f, "omitempty,alpha"); err != nil {
			return
		}
	}
	for _, f := range strings.Split(r.URL.Query().Get("location"), " ") {
		if err = validate.Var(f, "omitempty,alpha"); err != nil {
			return
		}
	}

	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 10
	}

	req = &Request{
		Name:     r.URL.Query().Get("name"),
		Type:     r.URL.Query().Get("type"),
		Field:    r.URL.Query().Get("field"),
		Language: r.URL.Query().Get("language"),
		Location: r.URL.Query().Get("location"),
		Page:     page,
		Limit:    limit,
	}
	err = validate.Struct(req)

	return req, err
}
