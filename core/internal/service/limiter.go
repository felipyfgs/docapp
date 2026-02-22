package service

import (
	"fmt"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.RWMutex
	limits   map[string]*bucket
	maxHits  int
	window   time.Duration
}

type bucket struct {
	hits      int
	resetAt   time.Time
	blockedAt *time.Time
}

func NewRateLimiter(maxHitsPerHour int) *RateLimiter {
	return &RateLimiter{
		limits:  make(map[string]*bucket),
		maxHits: maxHitsPerHour,
		window:  time.Hour,
	}
}

func (r *RateLimiter) Allow(cnpj string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	b, exists := r.limits[cnpj]
	if !exists {
		r.limits[cnpj] = &bucket{hits: 1, resetAt: now.Add(r.window)}
		return nil
	}

	if now.After(b.resetAt) {
		b.hits = 1
		b.resetAt = now.Add(r.window)
		b.blockedAt = nil
		return nil
	}

	if b.hits >= r.maxHits {
		if b.blockedAt == nil {
			b.blockedAt = &now
		}
		retryAfter := b.resetAt.Sub(now).Seconds()
		return fmt.Errorf("rate limit exceeded for %s, retry after %.0f seconds", cnpj, retryAfter)
	}

	b.hits++
	return nil
}

func (r *RateLimiter) MarkThrottled(cnpj string, retryAfterSeconds int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	resetAt := now.Add(time.Duration(retryAfterSeconds) * time.Second)

	b, exists := r.limits[cnpj]
	if !exists {
		r.limits[cnpj] = &bucket{
			hits:      r.maxHits,
			resetAt:   resetAt,
			blockedAt: &now,
		}
		return
	}

	b.hits = r.maxHits
	b.resetAt = resetAt
	b.blockedAt = &now
}

func (r *RateLimiter) RetryAfter(cnpj string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, exists := r.limits[cnpj]
	if !exists {
		return 0
	}

	now := time.Now()
	if now.After(b.resetAt) {
		return 0
	}

	return int(b.resetAt.Sub(now).Seconds())
}

func (r *RateLimiter) Reset(cnpj string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.limits, cnpj)
}

func (r *RateLimiter) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for cnpj, b := range r.limits {
		if now.After(b.resetAt) {
			delete(r.limits, cnpj)
		}
	}
}
