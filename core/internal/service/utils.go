package service

import (
	"math/rand"
	"strings"
	"time"
)

// firstNonEmpty returns the first string that is not entirely whitespace.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

// normalizeSiglaUF ensures the state code is a 2-letter uppercase string.
func normalizeSiglaUF(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	upper := strings.ToUpper(trimmed)
	if len(upper) > 2 {
		return upper[:2]
	}

	return upper
}

// sleepWithJitter blocks the current goroutine for a random duration between minSecs and maxSecs.
// It is useful to avoid thundering herd problems when making multiple API requests.
func sleepWithJitter(minSecs, maxSecs int) {
	jitter := time.Duration(minSecs)*time.Second +
		time.Duration(rand.Intn((maxSecs-minSecs)*1000+1))*time.Millisecond
	time.Sleep(jitter)
}
