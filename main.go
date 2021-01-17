package main

import (
	"github.com/gorilla/mux"
	"github.com/memochou1993/github-rankings/app/handler"
	"github.com/memochou1993/github-rankings/app/resource"
	"github.com/memochou1993/github-rankings/app/worker"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"log"
	"net/http"
)

func init() {
	util.LoadEnv()
	database.Init()
	logger.Init()
	resource.Init()
	worker.Init()
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ranking/owners", handler.ListOwners)
	r.HandleFunc("/ranking/repositories", handler.ListRepositories)
	log.Fatalln(http.ListenAndServe(":80", r))
}
