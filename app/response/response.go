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
	Cost      int64  `json:"cost,omitempty"`
	Limit     int64  `json:"limit,omitempty"`
	NodeCount int64  `json:"nodeCount,omitempty"`
	Remaining int64  `json:"remaining,omitempty"`
	ResetAt   string `json:"resetAt,omitempty"`
	Used      int64  `json:"used,omitempty"`
}

func (r RateLimit) Throttle(collecting int64) {
	logger.Debug(fmt.Sprintf("Rate Limit: %s", strconv.Quote(util.ParseStruct(r, " "))))
	resetAt, err := time.Parse(time.RFC3339, r.ResetAt)
	if err != nil {
		log.Fatal(err.Error())
	}
	remainingTime := resetAt.Add(time.Second).Sub(time.Now().UTC())
	time.Sleep(time.Duration(remainingTime.Milliseconds()/r.Remaining*collecting-500) * time.Millisecond)
	if r.Remaining > collecting {
		return
	}
	logger.Warning("Take a break...")
	time.Sleep(remainingTime)
}

type Error struct {
	Type      string `json:"type"`
	Locations []struct {
		Line   int64 `json:"line"`
		Column int64 `json:"column"`
	} `json:"locations"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}
