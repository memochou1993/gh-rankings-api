package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Response struct {
	Data struct {
		User struct {
			Repositories struct {
				Nodes []struct {
					Name string `json:"name"`
				} `json:"nodes"`
			} `json:"repositories"`
		} `json:"user"`
	} `json:"data"`
}

func main() {
	response, err := Fetch(context.Background())

	if err != nil {
		log.Println(err.Error())
	}

	fmt.Println(response)
}

func Fetch(ctx context.Context) (Response, error) {
	response := Response{}

	q := struct {
		Query string `json:"query"`
	}{
		Query: `
			query {
			  user(login: "memochou1993") {
				repositories(first: 10) {
				  nodes {
					name
				  }
				}
			  }
			}
		`,
	}

	body := &bytes.Buffer{}

	if err := json.NewEncoder(body).Encode(q); err != nil {
		return response, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.github.com/graphql", body)

	if err != nil {
		return response, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("ACCESS_TOKEN")))

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))

	if err != nil {
		return response, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err.Error())
		}
	}()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}
