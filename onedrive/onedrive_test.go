package onedrive

import (
	"strings"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		description string
		conf        Config
		// Substring the error message is intended to contain. Blank if no
		// error expected.
		errSubstr string
	}{
		{
			description: "valid config",
			conf: Config{
				TenantID:     "abc123abc123abc123",
				ClientID:     "abc123abc123abc123",
				ClientSecret: "abc123abc123abc123",
			},
			errSubstr: "",
		},
		{
			description: "missing client secret",
			conf: Config{
				TenantID: "abc123abc123abc123",
				ClientID: "abc123abc123abc123",
			},
			errSubstr: "client_secret",
		},
		{
			description: "missing client ID",
			conf: Config{
				TenantID:     "abc123abc123abc123",
				ClientSecret: "abc123abc123abc123",
			},
			errSubstr: "client_id",
		},
		{
			description: "missing tenant ID",
			conf: Config{
				ClientID:     "abc123abc123abc123",
				ClientSecret: "abc123abc123abc123",
			},
			errSubstr: "tenant_id",
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
			input:       "backups/2018-07-13 02:05.zip",
			want:        "backups/2018-07-13 02_05.zip",
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
			input:       "backups/this*is:\"a<bad>file?name\\to|use.zip",
			want:        "backups/this_is__a_bad_file_name_to_use.zip",
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
		{
			description: "Legal filename without a trailing .zip",
			input:       "myfile",
			want:        "myfile.zip",
			wantErr:     false,
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
