package dispatcher

import (
	"strconv"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

type RateLimitBucket struct {
	Remaining int
	Limit     int
	ResetAt   time.Time
}

type RateLimitMonitor struct {
	mu      sync.RWMutex
	buckets map[string]*RateLimitBucket
}

func NewRateLimitMonitor() *RateLimitMonitor {
	return &RateLimitMonitor{
		buckets: make(map[string]*RateLimitBucket),
	}
}

func (rlm *RateLimitMonitor) CanExecute(route string, guildID uint64) bool {
	key := rlm.getKey(route, guildID)

	rlm.mu.RLock()
	bucket, exists := rlm.buckets[key]
	rlm.mu.RUnlock()

	if !exists {
		return true
	}

	if time.Now().After(bucket.ResetAt) {
		return true
	}

	return bucket.Remaining > 0
}

func (rlm *RateLimitMonitor) UpdateFromFastHTTPResponse(resp *fasthttp.Response, route string, guildID uint64) {
	key := rlm.getKey(route, guildID)

	remaining := string(resp.Header.Peek("X-RateLimit-Remaining"))
	limit := string(resp.Header.Peek("X-RateLimit-Limit"))
	reset := string(resp.Header.Peek("X-RateLimit-Reset"))

	bucket := &RateLimitBucket{}

	if remaining != "" {
		bucket.Remaining, _ = strconv.Atoi(remaining)
	}
	if limit != "" {
		bucket.Limit, _ = strconv.Atoi(limit)
	}
	if reset != "" {
		resetUnix, _ := strconv.ParseInt(reset, 10, 64)
		bucket.ResetAt = time.Unix(resetUnix, 0)
	}

	rlm.mu.Lock()
	rlm.buckets[key] = bucket
	rlm.mu.Unlock()
}

func (rlm *RateLimitMonitor) getKey(route string, guildID uint64) string {
	return route + ":" + strconv.FormatUint(guildID, 10)
}

func (rlm *RateLimitMonitor) GetBucket(route string, guildID uint64) *RateLimitBucket {
	key := rlm.getKey(route, guildID)

	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	return rlm.buckets[key]
}
