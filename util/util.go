package util

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
)

func JoinStruct(v interface{}, sep string) string {
	b := bytes.Buffer{}
	encoder := json.NewEncoder(&b)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		log.Fatalln(err.Error())
	}

	s := b.String()
	s = strings.Replace(s, "\n", "", -1)
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	s = strings.Replace(s, "\\\"", "#", -1)
	s = strings.Replace(s, "\"", "", -1)
	s = strings.Replace(s, "#", "\"", -1)
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, ",", sep, -1)

	return s
}
