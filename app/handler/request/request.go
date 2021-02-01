package request

import (
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
)

type Request struct {
	Name  string `json:"name" validate:"omitempty,alphanum"`
	Type  string `json:"type" validate:"omitempty,alphanum"`
	Field string `json:"field" validate:"omitempty,alphanum"`
	Page  int64  `json:"page" validate:"omitempty,numeric"`
	Limit int64  `json:"limit" validate:"omitempty,numeric"`
}

func init() {
	validate = validator.New()
}

func (r *Request) IsNameEmpty() bool {
	return r.Name == ""
}

func (r *Request) IsTypeEmpty() bool {
	return r.Type == ""
}

func (r *Request) IsFieldEmpty() bool {
	return r.Field == ""
}

func (r *Request) IsEmpty() bool {
	return r.IsNameEmpty() && r.IsTypeEmpty() && r.IsFieldEmpty()
}

func Validate(r *http.Request) (req *Request, err error) {
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 10
	}
	req = &Request{
		Name:  r.URL.Query().Get("name"),
		Field: r.URL.Query().Get("field"),
		Type:  r.URL.Query().Get("type"),
		Page:  page,
		Limit: limit,
	}
	err = validate.Struct(req)
	return req, err
}
