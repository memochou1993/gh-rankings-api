package main

import (
	"github.com/gorilla/mux"
	"github.com/memochou1993/github-rankings/app/handler"
	"github.com/memochou1993/github-rankings/app/resource"
	"github.com/memochou1993/github-rankings/app/worker"
	"github.com/memochou1993/github-rankings/database"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/memochou1993/github-rankings/util"
	"net/http"
)

func init() {
	util.LoadEnv()
	database.Init()
	logger.Init()
	resource.Init()
}

func main() {
	w := worker.NewWorker()
	w.Init()

	r := mux.NewRouter()
	r.HandleFunc("/owners/{login}", handler.ShowOwner)
	http.ListenAndServe(":80", r)
}
