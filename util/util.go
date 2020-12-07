package util

import (
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
		log.Fatal(err)
	}

	env := os.Getenv("APP_ENV")
	if env != "" {
		env = fmt.Sprintf(".%s", env)
	}
	if err := godotenv.Load(fmt.Sprintf("%s/.env%s", dir, env)); err != nil {
		log.Fatal(err.Error())
	}
}

func LogStruct(name string, v interface{}) {
	log.Println(fmt.Sprintf("%s: \"%s\"", name, JoinStruct(v)))
}

func JoinStruct(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err.Error())
	}

	s := string(b)
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	s = strings.Replace(s, "\\\"", "_", -1)
	s = strings.Replace(s, "\"", "", -1)
	s = strings.Replace(s, "_", "\"", -1)

	return s
}
