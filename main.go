package main

import (
	"context"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/memochou1993/github-rankings/app"
	"log"
	"net/http"
	"os"
)

func main() {
	client := app.Client{
		Endpoint: "https://api.github.com/graphql",
		Token:    os.Getenv("ACCESS_TOKEN"),
		Client:   http.DefaultClient,
	}

	response, err := client.FetchUsers(context.Background())

	if err != nil {
		log.Println(err.Error())
	}

	fmt.Println(response)
}
