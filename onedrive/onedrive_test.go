package onedrive

import "testing"

func TestCleanFilename(t *testing.T) {
	makeLongFilename := func() string {
		len := 402
		var s string
		for i := 0; i < len; i++ {
			s += "i"
		}
		return s
	}

	cases := []struct {
		description string
		input       string
		want        string
		wantErr     bool
	}{
		{
			description: "Todoist backup name",
			input:       "backups/2018-07-13 02:05",
			want:        "backups/2018-07-13 02_05",
			wantErr:     false,
		},
		{
			description: "The filename is too long",
			input:       makeLongFilename(),
			want:        "",
			wantErr:     true,
		},
		{
			description: "Leading and trailing slashes",
			input:       "/backups/today.zip/",
			want:        "backups/today.zip",
			wantErr:     false,
		},
		{
			description: "Disallowed OneDrive characters",
			input:       "backups/this*is:\"a<bad>file?name\\to|use",
			want:        "backups/this_is__a_bad_file_name_to_use",
			wantErr:     false,
		},
		{
			description: "Disallowed file name",
			input:       "backups/todoist/CON",
			want:        "",
			wantErr:     true,
		},
		{
			description: "Disallowed folder name",
			input:       "backups/.lock/today",
			want:        "",
			wantErr:     true,
		},
		{
			description: "Disallowed folder name with number",
			input:       "backups/LPT3/today",
			want:        "",
			wantErr:     true,
		},
		{
			description: "Disallowed filename starting with ~$",
			input:       "backups/LPT3/~$today",
			want:        "",
			wantErr:     true,
		},
		{
			description: "Disallowed filename that includes _vti_ ~$",
			input:       "backups/LPT3/today_vti_",
			want:        "",
			wantErr:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			n, err := cleanFilename(tc.input)

			if (err != nil) != tc.wantErr {
				t.Fatalf(
					"wanted error status of %v but got %v with error %v",
					tc.wantErr,
					err != nil,
					err,
				)
			}

			if tc.want != n {
				t.Fatalf("wanted cleaned filename %q but got %q", tc.want, n)
			}
		})
	}
}
