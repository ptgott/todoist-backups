package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	azidentity "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/go-yaml/yaml"
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

const todoistBackupURL = "https://api.todoist.com/sync/v8/backups/get"

type AvailableBackups []AvailableBackup

type AvailableBackup struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}

// The OneDrive simple upload API supports uploads of up to 4MB.
// https://docs.microsoft.com/en-us/onedrive/developer/rest-api/api/driveitem_put_content
const oneDriveMaxBytes = 4e6

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

	// TODO: Validate the config

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
	t, err := cred.GetToken(ctx, shared.TokenRequestOptions{})

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
				t, err := cred.GetToken(ctx, shared.TokenRequestOptions{})

				if err != nil {
					fmt.Fprintln(os.Stderr, "Could not retrive an Azure AD auth token:", err.Error())
					os.Exit(1)
				}
			}

			tr1, err := http.NewRequest("GET", todoistBackupURL, nil)

			if err != nil {
				fmt.Fprintln(
					os.Stderr,
					"Unable to generate an HTTP request. URL:",
					todoistBackupURL,
					"Reason:",
					err.Error())
			}

			tr1.Header.Add("Authorization", "Bearer "+c.TodoistAPIKey)
			http.DefaultClient.Do(tr1)

			// TODO: Unmarshal the response into an AvailableBackup
			// TODO: Pick the latest available backup

			// TODO: Grab the latest available backup from Todoist
			// TODO: Send the payload to Microsoft Graph (grabbed the
			// token already--just need to build the URL path to send the payload to)
		case <-g:
			fmt.Fprintln(os.Stderr, "Received interrupt. Stopping.")
			os.Exit(0)
		}

	}
}
