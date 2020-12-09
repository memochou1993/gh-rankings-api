package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
)

func LoadEnv() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	if err := os.Chdir(dir); err != nil {
		log.Fatalln(err)
	}

	env := os.Getenv("APP_ENV")
	if env != "" {
		env = fmt.Sprintf(".%s", env)
	}
	if err := godotenv.Load(fmt.Sprintf("%s/.env%s", dir, env)); err != nil {
		log.Fatalln(err.Error())
	}
}

func Log(method string, v interface{}) {
	text := ""
	switch reflect.TypeOf(v).Kind() {
	case reflect.Struct:
		text = fmt.Sprintf("[%s] %s", method, JoinStruct(v, ", "))
	case reflect.String:
		text = fmt.Sprintf("[%s] %s", method, v)
	default:
		text = fmt.Sprintf("[%s] %v", method, v)
	}
	log.Println(text)
}

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
	s = strings.Replace(s, "\\\"", "_", -1)
	s = strings.Replace(s, "\"", "", -1)
	s = strings.Replace(s, "_", "\"", -1)
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, ",", sep, -1)

	return s
}
