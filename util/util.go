package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path"
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

func LogStruct(name string, v interface{}) {
	log.Println(fmt.Sprintf("%s: \"%s\"", name, JoinStruct(v)))
}

func JoinStruct(v interface{}) string {
	b := bytes.Buffer{}
	encoder := json.NewEncoder(&b)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		log.Fatalln(err.Error())
	}

	s := b.String()
	s = strings.Replace(s, "{", "", -1)
	s = strings.Replace(s, "}", "", -1)
	s = strings.Replace(s, "\\\"", "_", -1)
	s = strings.Replace(s, "\"", "", -1)
	s = strings.Replace(s, "_", "\"", -1)
	s = strings.Replace(s, "\n", "", -1)

	return s
}
