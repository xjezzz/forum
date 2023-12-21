package requestlimiter

import (
	"errors"
	"sync"
	"time"
)

type RateLimiter struct {
	mutex       sync.Mutex
	requestRate int
	lastRequest time.Time
}

func NewRateLimiter(requestRate int) *RateLimiter {
	return &RateLimiter{
		requestRate: requestRate,
	}
}

func (rl *RateLimiter) Limit() error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	currentTime := time.Now()
	timeSinceLastRequest := currentTime.Sub(rl.lastRequest)
	if timeSinceLastRequest < time.Second/time.Duration(rl.requestRate) {
		return errors.New("too many requests")
	}
	rl.lastRequest = currentTime
	return nil
}
