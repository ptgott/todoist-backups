package apiclient

import (
	"fmt"
	"net/http"
	"time"
)

type RetryConfig struct {
	IntervalBetweenRetries time.Duration
	MaxRetries             int
}

// DoWithRetries sends req, retrying on 5xx errors using the provided
// RetryConfig. Returns na error on non-2xx responses.
func DoWithRetries(c *http.Client, req *http.Request, f RetryConfig) (*http.Response, error) {
	remaining := f.MaxRetries

send:
	resp, err := c.Do(req)

	if err != nil {
		return nil, err
	}

	switch resp.StatusCode - (resp.StatusCode % 100) {
	// Retry in the case of server errors
	case 500:
		// We can retry, so wait a bit and try again.
		if remaining > 0 {
			remaining--
			time.Sleep(f.IntervalBetweenRetries)
			goto send
		}
		return resp, fmt.Errorf("the request to %v failed after %v retries", req.URL.String(), f.MaxRetries)
	case 400:
		return resp, fmt.Errorf("got client error %v", resp.StatusCode)
	case 200:
		return resp, nil
	default:
		return resp, fmt.Errorf("got unexpected response code %v", resp.StatusCode)
	}
}
