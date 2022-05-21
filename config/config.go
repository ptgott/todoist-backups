package config

import (
	"errors"
	"regexp"
)

type General struct {
	TodoistAPIKey string `yaml:"todoist_api_key"`

	// Must be a duration string like 1d or 3h
	BackupInterval string `yaml:"backup_interval"`
}

// Validate checks the config for errors
func (g General) Validate() error {
	if g.TodoistAPIKey == "" {
		return errors.New("must include a Todoist API key")
	}

	durRE := regexp.MustCompile("[1-9]+(ns|us|µs|ms|s|m|h)")
	if !durRE.MatchString(g.BackupInterval) {
		return errors.New("the backup interval must be a valid duration, e.g., 3h, as explained in: https://pkg.go.dev/time#Duration")
	}
	return nil
}
