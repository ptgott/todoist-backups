package todoist

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ptgott/todoist-backups/apiclient"
)

type AvailableBackups []AvailableBackup

type AvailableBackup struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}

const todoistBackupURL = "https://api.todoist.com/sync/v9/backups/get"

// The format Todoist uses for the timestamp of each backup. Numbers are
// assigned according to the conventions of the Go time package,
// https://pkg.go.dev/time
//
// See:
// https://developer.todoist.com/sync/v8/#get-backups
const todoistTimeFormat = "2006-01-02 15:04"

// GetAvailableBackups queries Todoist's sync API path for listing backups.
// It handles retries and returns an error either for a client issue or when
// all possibilities for retrieving available backups have been exhausted.
func GetAvailableBackups(token string) (AvailableBackups, error) {
	tr, err := http.NewRequest("GET", todoistBackupURL, nil)

	if err != nil {
		return AvailableBackups{},
			fmt.Errorf("unable to generate an HTTP request to %v:%v", todoistBackupURL, err)
	}

	tr.Header.Add("Authorization", "Bearer "+token)
	r, err := apiclient.DoWithRetries(
		http.DefaultClient,
		tr,
		apiclient.RetryConfig{
			IntervalBetweenRetries: time.Duration(10) * time.Minute,
			MaxRetries:             6,
		})

	if err != nil {
		return AvailableBackups{},
			fmt.Errorf("unexpected response while grabbing the latest Todoist backups: %v", err)
	}

	if r.StatusCode != 200 {
		return AvailableBackups{}, fmt.Errorf("got client error %v for URL %v", r.StatusCode, tr.URL)
	}

	var ab AvailableBackups
	// If this doesn't work, we can't proceed!
	if err := json.NewDecoder(r.Body).Decode(&ab); err != nil {
		return AvailableBackups{}, fmt.Errorf("unable to parse the available backups: %v", err)
	}

	return ab, nil
}

// GetBackup sends a GET request to the Todoist backup URL given in url with
// the provided bearer token. It writes the downloaded ZIP payload to w.
// Non-200 error codes will be returned as errors. If the payload reaches
// maxBytes in size, GetBackup will return an error.
func GetBackup(w io.Writer, token string, url string, maxBytes int64) error {
	tr, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return fmt.Errorf("unable to generate an HTTP request to %v:%v", url, err)
	}

	tr.Header.Add("Authorization", "Bearer "+token)
	r, err := apiclient.DoWithRetries(
		http.DefaultClient,
		tr,
		apiclient.RetryConfig{
			IntervalBetweenRetries: time.Duration(10) * time.Minute,
			MaxRetries:             6,
		})

	if err != nil {
		return fmt.Errorf("unexpected response while grabbing the latest Todoist backups: %v", err)
	}

	if r.StatusCode != 200 {
		return fmt.Errorf("got client error %v for URL %v", r.StatusCode, tr.URL)
	}

	lr := io.LimitReader(r.Body, maxBytes)
	i, err := io.Copy(w, lr)

	if i >= maxBytes {
		return errors.New("backup size exceeded upload limit")
	}

	return err
}

// LatestAvailableBackup returns a URL that callers can use to retrieve
// the latest Todoist backup.
func LatestAvailableBackup(ab AvailableBackups) (AvailableBackup, error) {
	if len(ab) == 0 {
		return AvailableBackup{}, errors.New("the list of available backups is empty")
	}

	var latestTime time.Time
	var latestAB AvailableBackup

	for _, t := range ab {
		if t.URL == "" {
			return AvailableBackup{}, errors.New("the list of possible backups includes a blank URL")
		}

		m, err := time.Parse(todoistTimeFormat, t.Version)

		if err != nil {
			return AvailableBackup{}, err
		}

		if m.After(latestTime) {
			latestTime = m
			latestAB = t
		}
	}
	return latestAB, nil
}
