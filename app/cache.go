package app

import (
	"github.com/patrickmn/go-cache"
	"time"
)

var (
	Cache = cache.New(24*time.Hour, 1*time.Hour)
)
