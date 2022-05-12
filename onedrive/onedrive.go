package onedrive

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

// See:
// https://docs.microsoft.com/en-us/onedrive/developer/rest-api/api/driveitem_put_content?view=odsp-graph-online#http-request-to-upload-a-new-file
const oneDriveUploadPath string = "https://graph.microsoft.com/me/drive/items/root:/%v:/content"

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
// returns an error if it is not possible to clean the filename.
func cleanFilename(filename string) (string, error) {
	// TODO: Add unit tests for this
	// TODO: look up OneDrive file naming restrictions
	// TODO: Remove leading slash and trailing slash

	// TODO: Return an error if the filename includes one of these characters:
	// " * : < > ? / \ |
	// See: https://support.microsoft.com/en-us/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb39e07630fa#invalidcharacters

	// TODO: Check each path segment for a disallowed file/folder name. These are listed here:
	// https://support.microsoft.com/en-us/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb39e07630fa#invalidfilefoldernames

	// TODO: The filepath must not be more than 400 characters
	// https://support.microsoft.com/en-us/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb39e07630fa#filenamepathlengths

	return "", nil

}
