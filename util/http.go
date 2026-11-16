package util

import (
	"net/http"
	"strconv"
	"time"
)

const (
	httpMaxAttempts = 8
	httpMaxWait     = 60 * time.Second
)

// httpBaseWait is the wait before the first retry. Tests shorten it.
var httpBaseWait = 2 * time.Second

// GetWithRetry sends a GET request and retries it while the server answers
// 429 Too Many Requests or a 5xx status, or the request fails to connect.
// The wait doubles from two seconds up to a minute, about three minutes over
// all attempts. A longer Retry-After header wins; a shorter one is ignored,
// because npm's rate limiter sends `Retry-After: 0` while it still refuses
// requests for several seconds. After the last attempt it returns that
// attempt's response or error, so callers must still check the status code.
func GetWithRetry(url string) (*http.Response, error) {
	wait := httpBaseWait
	for attempt := 1; ; attempt++ {
		resp, err := http.Get(url)
		if attempt == httpMaxAttempts || !shouldRetry(resp, err) {
			return resp, err
		}

		delay := wait
		if resp != nil {
			if after, ok := retryAfter(resp.Header.Get("Retry-After")); ok && after > delay {
				delay = after
			}
			resp.Body.Close()
		}
		if delay > httpMaxWait {
			delay = httpMaxWait
		}
		time.Sleep(delay)
		wait *= 2
	}
}

func shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
}

// retryAfter parses a Retry-After header, which is either a number of
// seconds or an HTTP date.
func retryAfter(value string) (time.Duration, bool) {
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second, true
	}
	if date, err := http.ParseTime(value); err == nil {
		return time.Until(date), true
	}
	return 0, false
}
