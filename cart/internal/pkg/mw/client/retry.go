package mw

import (
	"fmt"
	"net/http"
	"time"
)

const (
	retryDelay            = 1 * time.Second
	statusEnhanceYourCalm = 420
	statusTooManyRequests = 429
)

type RetryMiddleware struct {
	Transport  http.RoundTripper
	MaxRetries int
}

// RoundTrip function for retry request.
func (r *RetryMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i <= r.MaxRetries; i++ {
		resp, err = r.Transport.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		// If status not 420 or 429, return resp
		if resp.StatusCode != statusEnhanceYourCalm && resp.StatusCode != statusTooManyRequests {
			return resp, nil
		}

		resp.Body.Close()

		// If status 420 or 429, wait and retry
		if i < r.MaxRetries {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("max retries reached for status 420 or 429")
}
