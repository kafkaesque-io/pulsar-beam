package cache

import (
// "github.com/patrickmn/go-cache"
)

// This is an in memory cache for writekey validation.
// 1. Mostly read operation > 99.9999%
// 2. Small capacity < 1,000,000 rows
// 3. Highly concurrent > 500 goroutine

// Topic - Pulsar topic components
type Topic struct {
	tenant    string `json:"account"` // maps to Pulsar tenant
	namespace string `json:"namespace`
	topic     string `json:"topic"`
}

func init() {
	//where to initialize all DB connection
}

// WriteKeysCache - cache for writekeys
type WriteKeysCache struct {
}
