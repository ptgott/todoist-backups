package todoist

import (
	"encoding/json"
	"errors"
	"fmt"
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
	tr1, err := http.NewRequest("GET", todoistBackupURL, nil)

	if err != nil {
		return AvailableBackups{},
			fmt.Errorf("unable to generate an HTTP request to %v:%v", todoistBackupURL, err)
	}

	tr1.Header.Add("Authorization", "Bearer "+token)
	r1, err := http.DefaultClient.Do(tr1)

	// This error would likely be repeated on subsequent request
	// attempts. Bail out here so we can fix it.
	if err != nil {
		return AvailableBackups{},
			fmt.Errorf("unexpected response while grabbing the latest Todoist backups: %v", err)
	}

	if r1.StatusCode != 200 {
		// TODO: Add retries here
		return AvailableBackups{}, fmt.Errorf("got unexpected response %v", r1.StatusCode)
	}

	var ab AvailableBackups
	// If this doesn't work, we can't proceed!
	if err := json.NewDecoder(r1.Body).Decode(&ab); err != nil {
		return AvailableBackups{}, fmt.Errorf("unable to parse the available backups: %v", err)
	}

	return ab, nil
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
