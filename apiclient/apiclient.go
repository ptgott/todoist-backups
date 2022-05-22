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
// RetryConfig. It returns the response to the caller and does not return
// an error on non-2xx responses unless retrying has failed on a 5xx response.
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
	default:
		return resp, nil
	}
}
