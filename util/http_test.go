package util

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func shortenRetryWait(t *testing.T) {
	saved := httpBaseWait
	httpBaseWait = time.Millisecond
	t.Cleanup(func() { httpBaseWait = saved })
}

func TestGetWithRetryRetriesRateLimits(t *testing.T) {
	shortenRetryWait(t)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests <= 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, "error code: 1015")
			return
		}
		fmt.Fprint(w, `{"downloads":1}`)
	}))
	defer server.Close()

	resp, err := GetWithRetry(server.URL)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, requests)
}

func TestGetWithRetryReturnsOtherErrorsAtOnce(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resp, err := GetWithRetry(server.URL)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, 1, requests)
}

func TestGetWithRetryStopsAfterLastAttempt(t *testing.T) {
	shortenRetryWait(t)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	resp, err := GetWithRetry(server.URL)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Equal(t, httpMaxAttempts, requests)
}

func TestGetWithRetryWaitsDespiteRetryAfterZero(t *testing.T) {
	saved := httpBaseWait
	httpBaseWait = 200 * time.Millisecond
	t.Cleanup(func() { httpBaseWait = saved })

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	start := time.Now()
	resp, err := GetWithRetry(server.URL)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, time.Since(start) >= 200*time.Millisecond, "waited the backoff although Retry-After was 0")
}

func TestRetryAfter(t *testing.T) {
	d, ok := retryAfter("3")
	assert.True(t, ok)
	assert.Equal(t, 3*time.Second, d)

	d, ok = retryAfter(time.Now().Add(10 * time.Second).UTC().Format(http.TimeFormat))
	assert.True(t, ok)
	assert.InDelta(t, 10*time.Second, d, float64(2*time.Second))

	_, ok = retryAfter("")
	assert.False(t, ok)
	_, ok = retryAfter("soon")
	assert.False(t, ok)
}
