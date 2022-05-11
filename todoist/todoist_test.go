package todoist

import "testing"

func TestLatestAvailableBackup(t *testing.T) {
	cases := []struct {
		description string
		input       AvailableBackups
		expected    string
		expectErr   bool
	}{
		{
			description: "expected case",
			input: AvailableBackups{
				{
					Version: "2018-07-13 02:03",
					URL:     "https://downloads.example.com/1.zip",
				},
				{
					Version: "2018-07-13 02:04",
					URL:     "https://downloads.example.com/2.zip",
				},
				{
					Version: "2018-07-13 02:06",
					URL:     "https://downloads.example.com/4.zip",
				},
				{
					Version: "2018-07-13 02:05",
					URL:     "https://downloads.example.com/3.zip",
				},
			},
			expected:  "https://downloads.example.com/4.zip",
			expectErr: false,
		},
		{
			description: "one missing URL",
			input: AvailableBackups{
				{
					Version: "2018-07-13 02:03",
					URL:     "https://downloads.example.com/1.zip",
				},
				{
					Version: "2018-07-13 02:04",
					URL:     "",
				},
			},
			expected:  "",
			expectErr: true,
		},
		{
			description: "one missing version",
			input: AvailableBackups{
				{
					Version: "2018-07-13 02:03",
					URL:     "https://downloads.example.com/1.zip",
				},
				{
					Version: "",
					URL:     "https://downloads.example.com/2.zip",
				},
			},
			expected:  "",
			expectErr: true,
		},
		{
			description: "empty list",
			input:       AvailableBackups{},
			expected:    "",
			expectErr:   true,
		},
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
