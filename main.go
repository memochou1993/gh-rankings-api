package main

import (
	"github.com/gorilla/mux"
	"github.com/memochou1993/gh-rankings/app/handler"
	"github.com/memochou1993/gh-rankings/app/worker"
	"github.com/memochou1993/gh-rankings/database"
	"github.com/memochou1993/gh-rankings/util"
	"log"
	"net/http"
)

func init() {
	util.LoadEnv()
	database.Init()
	worker.Init()
}

func main() {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/", handler.Index).Methods(http.MethodGet)
	log.Fatal(http.ListenAndServe(":80", r))
}
