package gdrive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Config contains options that are necessary for using the Google Drive API
type Config struct {
	// Path to a credentials file, which you can export for a service account
	CredentialsPath string `yaml:"credentials_path"`
	// Path to a token file, which Google Workspace creates when you complete
	// the authorization flow.
	TokenPath string `yaml:"token_path"`
	// Folder where Todoist backups will be stored. This will be created at the
	// root of the Google Drive if it does not already exist.
	FolderName string `yaml:"folder_name"`
}

// Validate checks the Config for errors and returns the first one it finds.
func (c Config) Validate() error {
	if c.CredentialsPath == "" || c.TokenPath == "" || c.FolderName == "" {
		return errors.New("must provide a folder_name, credentials_path, and token_path")
	}

	if _, err := os.Stat(c.CredentialsPath); err != nil {
		return errors.New("cannot find a file at credentials_path")
	}

	if _, err := os.Stat(c.TokenPath); err != nil {
		return errors.New("cannot find a file at token_path")
	}

	return nil
}

// Returns a Google Drive API client based on the token at the provided path.
//
// Per the Google Drive Golang quickstart guide:
// "The token file stores the user's access and refresh tokens, and is
// created automatically when the authorization flow completes for the first
// time." (https://developers.google.com/drive/api/quickstart/go)
func getClient(config *oauth2.Config, path string) (*http.Client, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	if err = json.NewDecoder(f).Decode(tok); err != nil {
		return nil, err
	}
	return config.Client(context.Background(), tok), nil
}

// UploadFile uploads the file in r to Google Drive with the provided name.
// It modifies the filename to remove invalid characters before uploading.
//
// UploadFile creates a Google Drive API client using the token file at tokPath
// and credentials file at credPath.
func UploadFile(r io.Reader, filename string, c Config) error {
	ctx := context.Background()
	b, err := ioutil.ReadFile(c.CredentialsPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// Request the DriveFileScope so this app can only interact with files it
	// creates.
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client, err := getClient(config, c.TokenPath)

	if err != nil {
		return err
	}

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	var d string // the ID of the directory to write to

	l, err := srv.Files.List().Q(fmt.Sprintf("name='%v'", c.FolderName)).Do()

	if err != nil {
		return err
	}

	switch len(l.Files) {
	case 0:
		// Create a folder for your backups since it does not already exist
		s, err := srv.Files.Create(
			&drive.File{
				Name: c.FolderName,
				// Indicate that this is a new folder. See:
				// https://developers.google.com/drive/api/guides/folder#create_a_folder
				MimeType: "application/vnd.google-apps.folder",
			},
		).Context(ctx).Do()

		if err != nil {
			return err
		}

		d = s.Id
	case 1:
		// Use the ID of the existing folder
		d = l.Files[0].Id
	default:
		return fmt.Errorf(
			"unexpected number of Todoist backup folders: %v files named %q",
			len(l.Files),
			c.FolderName,
		)
	}

	if _, err := srv.Files.Create(&drive.File{
		MimeType: "application/zip",
		Name:     filename,
		Parents:  []string{d},
	}).Media(r).Context(ctx).Do(); err != nil {
		return err
	}

	return nil
}
