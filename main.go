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
	"github.com/ptgott/todoist-backups/config"
	"github.com/ptgott/todoist-backups/onedrive"
	"github.com/ptgott/todoist-backups/todoist"
)

type Config struct {
	config.General
	OneDrive onedrive.Config `json:"onedrive"`
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

func runBackup(cred *azidentity.ClientSecretCredential, c Config) {
	// Ensure that the OneDrive credentials are scoped only to the given
	// directory.
	//
	// To do this, request authorization for the "Files.ReadWrite.AppFolder"
	// scope. The user authorizes the app to access their app folder.
	//
	// App folders are only compatible with personal OneDrive accounts.
	// https://docs.microsoft.com/en-us/onedrive/developer/rest-api/concepts/special-folders-appfolder
	ctx := context.Background()
	t, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{
			// Added via the SDK:
			// https://github.com/microsoft/kiota-authentication-azure-go/blob/474cb0d2c8b20401adf95c1d359c59ba4fe565b6/azure_identity_access_token_provider.go#L40
			"https://graph.microsoft.com/.default",
			"https://graph.microsoft.com/Files.ReadWrite.AppFolder",
		},
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not retrive an Azure AD auth token:", err.Error())
		os.Exit(1)
	}

	ab, err := todoist.GetAvailableBackups(c.TodoistAPIKey)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to grab the available backups from Todoist:", err.Error())
		os.Exit(1)
	}

	u, err := todoist.LatestAvailableBackup(ab)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to determine the latest available backup from Todoist:", err.Error())
		os.Exit(1)
	}

	var buf bytes.Buffer
	if err := todoist.GetBackup(&buf, c.TodoistAPIKey, u.URL, oneDriveMaxBytes); err != nil {
		fmt.Fprintln(os.Stderr, "Unable to retrieve the latest Todoist backup:", err.Error())
		os.Exit(1)
	}

	if err := onedrive.UploadFile(&buf, t, c.OneDrive.DirectoryPath+"/"+u.Version); err != nil {
		fmt.Fprintln(os.Stderr, "Unable to upload a file to OneDrive", err.Error())
		os.Exit(1)
	}
}

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

	if err := c.General.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, "Invalid config: "+err.Error())
		os.Exit(1)
	}

	if err := c.OneDrive.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, "Invalid OneDrive config: "+err.Error())
		os.Exit(1)
	}

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

	dur, err := time.ParseDuration(c.BackupInterval)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not parse the backup interval:", err.Error())
		os.Exit(1)
	}

	// Run the first backup right away so we can identify issues
	runBackup(cred, c)

	k := time.NewTicker(dur)
	for {
		select {
		case <-k.C:
			runBackup(cred, c)
		case <-g:
			fmt.Fprintln(os.Stderr, "Received interrupt. Stopping.")
			os.Exit(0)
		}
	}
}
