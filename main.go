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
	"github.com/rs/zerolog/log"
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
		log.Fatal().Err(err).Msg("Could not retrive an Azure AD auth token")
	}

	ab, err := todoist.GetAvailableBackups(c.TodoistAPIKey)

	if err != nil {
		log.Fatal().Err(err).Msg("Unable to grab the available backups from Todoist")
	}

	u, err := todoist.LatestAvailableBackup(ab)

	if err != nil {
		log.Fatal().Err(err).Msg("Unable to determine the latest available backup from Todoist")
	}

	var buf bytes.Buffer
	if err := todoist.GetBackup(&buf, c.TodoistAPIKey, u.URL, oneDriveMaxBytes); err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve the latest Todoist backup")
	}

	if err := onedrive.UploadFile(&buf, t, c.OneDrive.DirectoryPath+"/"+u.Version); err != nil {
		log.Fatal().Err(err).Msg("Unable to upload a file to OneDrive")
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
		log.Fatal().Str("filepath", *cf).Msg("Could not open the config file:")
	}

	var c Config

	if err := yaml.NewDecoder(f).Decode(&c); err != nil {
		log.Fatal().Err(err).Msg("Could not parse your config file")
	}

	if err := c.General.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid config")
	}

	if err := c.OneDrive.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid OneDrive config")
	}

	cred, err := azidentity.NewClientSecretCredential(
		c.OneDrive.TenantID,
		c.OneDrive.ClientID,
		c.OneDrive.ClientSecret,
		nil,
	)

	if err != nil {
		log.Fatal().Err(err).Msg("Could not authenticate with Azure AD")
	}

	dur, err := time.ParseDuration(c.BackupInterval)

	if err != nil {
		log.Fatal().Err(err).Msg("Could not parse the backup interval")
	}

	// Run the first backup right away so we can identify issues
	log.Info().Msg("running initial backup")
	runBackup(cred, c)

	k := time.NewTicker(dur)
	for {
		select {
		case <-k.C:
			log.Info().Msg("running periodic backup")
			runBackup(cred, c)
		case <-g:
			log.Info().Msg("Received interrupt. Stopping.")
			os.Exit(0)
		}
	}
}
