package request

import (
	"github.com/go-playground/validator/v10"
	"strings"
)

var (
	validate *validator.Validate
)

func init() {
	validate = validator.New()
}

func sanitize(text string) string {
	symbols := []string{"@", "$", "%", "^", "&", "[", "]", "{", "}", "<", ">"}
	for _, symbol := range symbols {
		text = strings.Trim(text, symbol)
	}
	return text
}
