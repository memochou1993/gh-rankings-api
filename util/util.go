package util

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path"
	"runtime"
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
