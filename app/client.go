package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/memochou1993/github-rankings/app/collection"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Client struct {
	Endpoint string
	Token    string
	Client   *http.Client
}

type Query struct {
	Query string `json:"query"`
}

func (c *Client) FetchUsers(ctx context.Context) (collection.Users, error) {
	response := collection.Users{}
	err := c.fetch(ctx, c.readQuery("users"), &response)

	return response, err
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
	req, err := http.NewRequest(http.MethodPost, c.Endpoint, body)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	return c.Client.Do(req.WithContext(ctx))
}

func (c *Client) readQuery(name string) []byte {
	data, err := ioutil.ReadFile(fmt.Sprintf("./app/collection/%s.graphql", name))

	if err != nil {
		log.Fatal(err.Error())
	}

	return data
}
