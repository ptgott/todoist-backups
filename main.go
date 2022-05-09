package main

import (
	"flag"
	"fmt"
	"os"
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
4h, or 3d.

`

func main() {
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

	// TODO: Decode the YAML into a Config
	// TODO: Validate the config
	// TODO: in a loop:
	// TODO: Authenticate with Microsoft Graph/Azure AD
	// TODO: Get a list of backups to download from Todoist
	// TODO: Grab a payload from Todoist
	// TODO: Send the payload to Microsoft Graph
}
