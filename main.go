package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	azidentity "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/go-yaml/yaml"
	"github.com/ptgott/todoist-backups/onedrive"
	"github.com/ptgott/todoist-backups/todoist"
)

type Config struct {
	TodoistAPIKey string         `json:"todoist_api_key"`
	OneDrive      OneDriveConfig `json:"onedrive"`
	// Must be a duration string like 1d or 3h
	BackupInterval string `json:"backup_interval"`
}

type OneDriveConfig struct {
	TenantID      string `json:"tenant_id"`
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	DirectoryPath string `json:"directory_path"`
}

// The OneDrive simple upload API supports uploads of up to 4MB.
// https://docs.microsoft.com/en-us/onedrive/developer/rest-api/api/driveitem_put_content
const oneDriveMaxBytes int64 = 4e6

// For LimitReaders: 5MB
const maxResponseBodyBytes int64 = 5e6

const help = `You must provide a -config flag with the path to a config file.

The config file must include the following options in YAML format:

todoist_api_key: the API key retrieved from Todoist

onedrive:
	tenant_id: Microsoft Graph tenant ID

	client_id: Microsoft Graph client ID

	client_secret: Microsoft Graph client secret

	directory_path: path to the OneDrive directory you want to write backups
	to.

	The Todoist backup job will be limited to this directory.

backup_interval: How often to conduct the backup. A duration string like 1m, 
4h, or 3d.`

func main() {
	var g chan os.Signal
	signal.Notify(g, os.Interrupt)

	cf := flag.String("config", "", "the path to a configuration file")
	flag.Parse()

	if *cf == "" {
		fmt.Println(help)
		os.Exit(1)
	}

	f, err := os.Open(*cf)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not open the config file:", cf)
		os.Exit(1)
	}

	var c Config

	if err := yaml.NewDecoder(f).Decode(&c); err != nil {
		fmt.Fprintln(os.Stderr, "Could not parse your config file: "+err.Error())
		os.Exit(1)
	}

	// TODO: Validate the config.

	cred, err := azidentity.NewClientSecretCredential(
		c.OneDrive.TenantID,
		c.OneDrive.ClientID,
		c.OneDrive.ClientSecret,
		nil,
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not authenticate with Azure AD:", err.Error())
		os.Exit(1)
	}

	// Grab a token before entering the main loop to flag any authentication
	// issues early.
	ctx := context.Background()
	t, err := cred.GetToken(ctx, policy.TokenRequestOptions{})

	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not retrive an Azure AD auth token:", err.Error())
		os.Exit(1)
	}

	dur, err := time.ParseDuration(c.BackupInterval)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not parse the backup interval:", err.Error())
		os.Exit(1)
	}
	// TODO: Execute a GET request against the Todoist API before the main loop
	// to help troubleshoot

	k := time.NewTicker(dur)

	for {
		select {
		case <-k.C:
			if t.ExpiresOn.After(time.Now()) {
				t, err := cred.GetToken(ctx, policy.TokenRequestOptions{})

				if err != nil {
					fmt.Fprintln(os.Stderr, "Could not retrive an Azure AD auth token:", err.Error())
					os.Exit(1)
				}
			}

			ab, err := todoist.GetAvailableBackups(c.TodoistAPIKey)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to grab the available backups from Todoist:", err.Error())
				os.Exit(1)
			}

			u, err := todoist.LatestAvailableBackup(ab)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to determine the latest available backup from Todoist:", err.Error())
				os.Exit(1)
			}

			var buf bytes.Buffer
			if err := todoist.GetBackup(&buf, c.TodoistAPIKey, u.URL, oneDriveMaxBytes); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to retrieve the latest Todoist backup:", err.Error())
				os.Exit(1)
			}

			if err := onedrive.UploadFile(&buf, t, u.Version); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to upload a file to OneDrive", err.Error())
				os.Exit(1)
			}

		case <-g:
			fmt.Fprintln(os.Stderr, "Received interrupt. Stopping.")
			os.Exit(0)
		}

	}
}
