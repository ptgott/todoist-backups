package onedrive

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

type Config struct {
	TenantID     string `json:"tenant_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// Validate checks the config for errors
func (c Config) Validate() error {
	fields := map[string]string{
		c.ClientSecret: "client_secret",
		c.ClientID:     "client_id",
		c.TenantID:     "tenant_id",
	}

	for k, v := range fields {
		if k == "" {
			return errors.New("the OneDrive config must include the field: " + v)
		}
	}
	return nil
}

// Upload path to use for new content
// See:
// https://docs.microsoft.com/en-us/onedrive/developer/rest-api/api/driveitem_put_content?view=odsp-graph-online#http-request-to-upload-a-new-file
// We are creating an App Folder, so we need to specify this URL path.
// See:
// https://docs.microsoft.com/en-us/onedrive/developer/rest-api/concepts/special-folders-appfolder#creating-your-apps-folder
const oneDriveUploadPath string = "/drive/special/approot:/%v:/content"

// UploadFile sends a request to the OneDrive API to upload the file in body.
// Filename must be relative to the root of your OneDrive file tree, and must
// not have a leading "/". The file will be created with filename, but
// modifications may be made to accommodate OneDrive's policies.
//
// No validation is performed on body before uploading.
func UploadFile(body io.Reader, k *azcore.AccessToken, filename string) error {
	fn, err := cleanFilename(filename)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf(oneDriveUploadPath, fn), body)

	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+k.Token)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		return errors.New("got unexpected response code: " + strconv.Itoa(resp.StatusCode))
	}

	return nil

}

// cleanFilename modifies filename for use in OneDrive API requests, and
// returns an error if it is not possible to clean the filename. Path segments
// must be separated with "/" characters, not "\" characters. Each illegal
// character is replaced with a "_".
func cleanFilename(filename string) (string, error) {
	// The filepath must not be more than 400 characters
	// https://support.microsoft.com/en-us/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb39e07630fa#filenamepathlengths
	if len(filename) > 400 {
		return "", errors.New("the filepath cannot exceed 400 characters")
	}

	var s string

	// Remove leading slash and trailing slash
	s = strings.Trim(filename, "/")

	// Replace illegal OneDrive filename characters with underscores
	// The characters are:
	// " * : < > ? / \ |
	// See: https://support.microsoft.com/en-us/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb39e07630fa#invalidcharacters
	// However, note that we are treating "/" characters as path separators.
	s = regexp.MustCompile("[\"*:<>?\\\\|]").ReplaceAllString(s, "_")
	_, fn := path.Split(filename)
	if strings.Contains(fn, "~$") {
		return "", errors.New("the filename cannot include \"~$\"")
	}
	if strings.Contains(fn, "_vti_") {
		return "", errors.New("the filename cannot include \"_vti_\"")
	}

	// Check each path segment for a disallowed file/folder name. These are listed here:
	// https://support.microsoft.com/en-us/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb39e07630fa#invalidfilefoldernames
	bn := regexp.MustCompile("^(\\.lock|CON|PRN|AUX|NUL|COM[0-9]|LPT[0-9]|_vti_|desktop\\.ini)$")
	for _, p := range strings.Split(filename, "/") {
		if bn.MatchString(p) {
			return "", errors.New("filepath contains disallowed file/folder name: " + p)
		}
	}

	return s, nil

}
