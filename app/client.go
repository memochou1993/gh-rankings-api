package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/memochou1993/github-rankings/app/model"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Client struct {
	Client *http.Client
}

type Query struct {
	Query string `json:"query"`
}

func (c *Client) GetClient() *http.Client {
	if c.Client != nil {
		return c.Client
	}

	c.Client = http.DefaultClient

	return c.Client
}

func (c *Client) SearchUsers(ctx context.Context) (model.SearchedUsers, error) {
	data := model.SearchedUsers{}
	err := c.fetch(ctx, c.readQuery("search_users"), &data)

	return data, err
}

func (c *Client) fetch(ctx context.Context, q []byte, v interface{}) error {
	body := &bytes.Buffer{}

	if err := json.NewEncoder(body).Encode(Query{Query: string(q)}); err != nil {
		return err
	}

	resp, err := c.post(ctx, body)

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err.Error())
		}
	}()

	if err != nil {
		return err
	}

	return json.NewDecoder(resp.Body).Decode(&v)
}

func (c *Client) post(ctx context.Context, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, os.Getenv("API_URL"), body)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("API_TOKEN")))

	return c.GetClient().Do(req.WithContext(ctx))
}

func (c *Client) readQuery(name string) []byte {
	data, err := ioutil.ReadFile(fmt.Sprintf("./app/query/%s.graphql", name))

	if err != nil {
		log.Fatal(err.Error())
	}

	return data
}
