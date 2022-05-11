package todoist

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AvailableBackups []AvailableBackup

type AvailableBackup struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}

const todoistBackupURL = "https://api.todoist.com/sync/v8/backups/get"

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
	r, err := http.DefaultClient.Do(tr)

	// This error would likely be repeated on subsequent request
	// attempts. Bail out here so we can fix it.
	if err != nil {
		return AvailableBackups{},
			fmt.Errorf("unexpected response while grabbing the latest Todoist backups: %v", err)
	}

	if r.StatusCode != 200 {
		// TODO: Add retries here
		return AvailableBackups{}, fmt.Errorf("got unexpected response %v", r.StatusCode)
	}

	var ab AvailableBackups
	// If this doesn't work, we can't proceed!
	if err := json.NewDecoder(r.Body).Decode(&ab); err != nil {
		return AvailableBackups{}, fmt.Errorf("unable to parse the available backups: %v", err)
	}

	return ab, nil
}

// GetBackup sends a GET request to the Todoist backup URL given in url with
// the provided bearer token. It writes the downloaded ZIP payload to w and
// returns the number of bytes downloaded along with any errors. Non-200 error
// codes will be returned as errors. If the payload reaches maxBytes in size,
// GetBackup will return an error.
func GetBackup(w io.Writer, token string, url string, maxBytes int64) (int64, error) {
	tr, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return 0,
			fmt.Errorf("unable to generate an HTTP request to %v:%v", url, err)
	}

	tr.Header.Add("Authorization", "Bearer "+token)
	r, err := http.DefaultClient.Do(tr)

	// This error would likely be repeated on subsequent request
	// attempts. Bail out here so we can fix it.
	if err != nil {
		return 0,
			fmt.Errorf("unexpected response while grabbing the latest Todoist backups: %v", err)
	}

	if r.StatusCode != 200 {
		// TODO: Add retries here
		return 0, fmt.Errorf("got unexpected response %v", r.StatusCode)
	}

	lr := io.LimitReader(r.Body, maxBytes)
	i, err := io.Copy(w, lr)

	if i >= maxBytes {
		return i, errors.New("backup size exceeded OneDrive upload limit")
	}

	return i, err
}

// LatestAvailableBackup returns a URL that callers can use to retrieve
// the latest Todoist backup.
func LatestAvailableBackup(ab AvailableBackups) (string, error) {
	if len(ab) == 0 {
		return "", errors.New("the list of available backups is empty")
	}

	var latestTime time.Time
	var latestURL string

	for _, t := range ab {
		if t.URL == "" {
			return "", errors.New("the list of possible backups includes a blank URL")
		}

		m, err := time.Parse(todoistTimeFormat, t.Version)

		if err != nil {
			return "", err
		}

		if m.After(latestTime) {
			latestTime = m
			latestURL = t.URL
		}
	}
	return latestURL, nil
}
