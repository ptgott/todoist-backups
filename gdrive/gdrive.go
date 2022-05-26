package gdrive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Config contains options that are necessary for using the Google Drive API
type Config struct {
	// Path to a credentials file, which you can export for a service account.
	// Must be a service account credentials file, not an OAuth2.0 token file.
	CredentialsPath string `yaml:"credentials_path"`
	// Folder where Todoist backups will be stored. This will be created at the
	// root of the Google Drive if it does not already exist.
	FolderName string `yaml:"folder_name"`
}

// Validate checks the Config for errors and returns the first one it finds.
func (c Config) Validate() error {
	if c.CredentialsPath == "" || c.FolderName == "" {
		return errors.New("must provide a folder_name and credentials_path")
	}

	if _, err := os.Stat(c.CredentialsPath); err != nil {
		return errors.New("cannot find a file at credentials_path")
	}

	return nil
}

// UploadFile uploads the file in r to Google Drive with the provided name.
// The containing folder (Config.FolderName) must exist and be shared
// with the Todoist backupos service account prior to the upload.
func UploadFile(r io.Reader, filename string, c Config) error {
	ctx := context.Background()

	srv, err := drive.NewService(ctx, option.WithCredentialsFile(c.CredentialsPath))
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
		return fmt.Errorf("could not find backup folder %q", c.FolderName)
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
