package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/memochou1993/github-rankings/logger"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var client *http.Client

func init() {
	client = newClient()
}

func newClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

func Fetch(ctx context.Context, q string, v interface{}) error {
	body := strings.NewReader(q)
	resp, err := post(ctx, body)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatalln(err.Error())
		}
	}()
	if resp.StatusCode != http.StatusOK {
		logger.Error("Something went wrong...")
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func post(ctx context.Context, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, viper.GetString("API_URL"), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", viper.GetString("API_TOKEN")))
	return client.Do(req.WithContext(ctx))
}
