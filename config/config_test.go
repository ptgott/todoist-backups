package config

import (
	"strings"
	"testing"
)

func TestGeneralValidate(t *testing.T) {
	cases := []struct {
		description string
		conf        General
		// Substring the error message is intended to contain. Blank if no
		// error expected.
		errSubstr string
	}{
		{
			description: "valid case",
			conf: General{
				TodoistAPIKey:  "123abc123abc123abc",
				BackupInterval: "3h",
			},
			errSubstr: "",
		},
		{
			description: "backup interval has no unit",
			conf: General{
				TodoistAPIKey:  "123abc123abc123abc",
				BackupInterval: "3",
			},
			errSubstr: "duration",
		},
		{
			description: "backup interval has no number",
			conf: General{
				TodoistAPIKey:  "123abc123abc123abc",
				BackupInterval: "h",
			},
			errSubstr: "duration",
		},
		{
			description: "backup interval has an unsupported duration unit",
			conf: General{
				TodoistAPIKey:  "123abc123abc123abc",
				BackupInterval: "1y",
			},
			errSubstr: "duration",
		},
		{
			description: "missing backup interval",
			conf: General{
				TodoistAPIKey: "123abc123abc123abc",
			},
			errSubstr: "backup interval",
		},
		{
			description: "missing API key",
			conf: General{
				BackupInterval: "3h",
			},
			errSubstr: "API key",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.conf.Validate()

			if err != nil && tc.errSubstr == "" {
				t.Fatalf("expected no error but got %v", err)
			}

			if err == nil && tc.errSubstr != "" {
				t.Fatal("expected an error but got nil")
			}

			if err == nil && tc.errSubstr == "" {
				return
			}

			if !strings.Contains(err.Error(), tc.errSubstr) {
				t.Fatalf("could not find expected substring %q in error message %q", tc.errSubstr, err.Error())
			}
		})
	}
}
