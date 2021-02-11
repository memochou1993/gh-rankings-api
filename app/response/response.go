package response

import (
	"fmt"
	"github.com/memochou1993/gh-rankings/logger"
	"github.com/memochou1993/gh-rankings/util"
	"log"
	"strconv"
	"time"
)

type PageInfo struct {
	EndCursor   string `json:"endCursor,omitempty"`
	HasNextPage bool   `json:"hasNextPage,omitempty"`
}

type RateLimit struct {
	Cost      int    `json:"cost,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	NodeCount int    `json:"nodeCount,omitempty"`
	Remaining int    `json:"remaining,omitempty"`
	ResetAt   string `json:"resetAt,omitempty"`
	Used      int    `json:"used,omitempty"`
}

func (r RateLimit) Break(collecting int) {
	period := float64(3600) / float64(5000)
	time.Sleep(time.Duration(period*float64(collecting)*1000) * time.Millisecond)
	logger.Debug(fmt.Sprintf("Rate Limit: %s", strconv.Quote(util.ParseStruct(r, " "))))

	buffer := 10
	if r.Remaining > buffer {
		return
	}
	resetAt, err := time.Parse(time.RFC3339, r.ResetAt)
	if err != nil {
		log.Fatal(err.Error())
	}

	logger.Warning("Take a break...")
	time.Sleep(resetAt.Add(time.Second).Sub(time.Now().UTC()))
}

type Error struct {
	Type      string `json:"type"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}
