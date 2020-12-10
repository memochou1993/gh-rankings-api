package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/memochou1993/github-rankings/util"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var client *http.Client

func init() {
	util.LoadEnv()
	initClient()
}

func initClient() {
	client = http.DefaultClient
}

func Query(ctx context.Context, r *Request, v interface{}) error {
	body := strings.NewReader(r.Query())
	resp, err := post(ctx, body)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatalln(err.Error())
		}
	}()

	return json.NewDecoder(resp.Body).Decode(v)
}

func post(ctx context.Context, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, os.Getenv("API_URL"), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("API_TOKEN")))

	return client.Do(req.WithContext(ctx))
}
