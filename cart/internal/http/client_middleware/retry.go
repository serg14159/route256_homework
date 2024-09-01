package client_middleware

import (
	"fmt"
	"net/http"
	"time"
)

type RetryMiddleware struct {
	Transport  http.RoundTripper
	MaxRetries int
}

// Function for retry request.
func (r *RetryMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.Transport == nil {
		r.Transport = http.DefaultTransport
	}

	var resp *http.Response
	var err error

	for i := 0; i <= r.MaxRetries; i++ {
		resp, err = r.Transport.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		// If status not 420 or 429, return resp
		if resp.StatusCode != 420 && resp.StatusCode != 429 {
			return resp, nil
		}

		resp.Body.Close()

		// If status 420 or 429, wait and retry
		if i < r.MaxRetries {
			time.Sleep(1 * time.Second)
		}
	}

	return nil, fmt.Errorf("max retries reached for status 420 or 429")
}
