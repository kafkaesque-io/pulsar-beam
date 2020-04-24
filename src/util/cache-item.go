package util

import (
	"time"
)

const (
	// ItemNotExpire avoids the item being expired by TTL
	ItemNotExpire time.Duration = -1
)

func newItem(key string, object interface{}, ttl time.Duration) *item {
	item := &item{
		data: object,
		ttl:  ttl,
		key:  key,
	}
	// no mutex is required for the first time
	item.touch()
	return item
}

type item struct {
	key      string
	data     interface{}
	ttl      time.Duration
	expireAt time.Time
}

// Reset the item expiration time
func (item *item) touch() {
	if item.ttl > 0 {
		item.expireAt = time.Now().Add(item.ttl)
	}
}

// Verify if the item is expired
func (item *item) expired() bool {
	if item.ttl <= 0 {
		return false
	}
	return item.expireAt.Before(time.Now())
}
