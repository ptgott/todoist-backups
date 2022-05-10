package todoist

import "testing"

func TestLatestAvailableBackup(t *testing.T) {
	cases := []struct {
		description string
		input       AvailableBackups
		expected    string
		expectErr   bool
	}{
		// TODO: Add test cases
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			b, err := LatestAvailableBackup(tc.input)

			if (err != nil) != tc.expectErr {
				t.Fatalf("expected error status of %v but got %v with error %v", tc.expectErr, err != nil, err)
			}

			if b != tc.expected {
				t.Fatalf("expected latest backup URL %v but got %v", tc.expected, b)
			}
		})
	}
}
