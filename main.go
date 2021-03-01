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
	database.Connect()
	worker.Start()
}

func main() {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ranks", handler.ListRanks).Methods(http.MethodGet)
	api.HandleFunc("/users", handler.ListUsers).Methods(http.MethodGet)
	api.HandleFunc("/users/{login}", handler.ShowUser).Methods(http.MethodGet)
	api.HandleFunc("/organizations", handler.ListOrganizations).Methods(http.MethodGet)
	api.HandleFunc("/organizations/{login}", handler.ShowOrganization).Methods(http.MethodGet)
	api.HandleFunc("/repositories", handler.ListRepositories).Methods(http.MethodGet)
	api.HandleFunc("/repositories/{owner}/{name}", handler.ShowRepository).Methods(http.MethodGet)
	log.Fatal(http.ListenAndServe(":80", r))
}
