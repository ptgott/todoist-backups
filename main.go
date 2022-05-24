package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/ptgott/todoist-backups/config"
	"github.com/ptgott/todoist-backups/gdrive"
	"github.com/ptgott/todoist-backups/todoist"
	"github.com/rs/zerolog/log"
)

type Config struct {
	General     config.General `yaml:"general"`
	GoogleDrive gdrive.Config  `yaml:"google_drive"`
}

// For LimitReaders: 5MB
const maxResponseBodyBytes int64 = 5e6

const help = `
You must provide a -config flag with the path to a config file.

The config file must include the following options in YAML format:

general:

	todoist_api_key: the API key retrieved from Todoist

	backup_interval: How often to conduct the backup. A duration string like 1m, 
	4h, or 3d.

google_drive:
	token_path: path to your Google Workspace token file, which is created when
	you first complete the authorization flow.

	credentials_path: path to a Google Workspace credentials file, which you
	can export for the service account that you created for this app.

	folder_name: name of the Google Drive directory you want to write 
	backups to. This will be a single folder at the root of your Drive.

	The Todoist backup job will be limited to this directory.

You can optionally use the -oneshot flag to create a single backup without
running the job as a daemon.
`

func runBackup(c Config) {
	ab, err := todoist.GetAvailableBackups(c.General.TodoistAPIKey)

	if err != nil {
		log.Fatal().Err(err).Msg("Unable to grab the available backups from Todoist")
	}

	u, err := todoist.LatestAvailableBackup(ab)

	if err != nil {
		log.Fatal().Err(err).Msg("Unable to determine the latest available backup from Todoist")
	}

	var buf bytes.Buffer
	if err := todoist.GetBackup(&buf, c.General.TodoistAPIKey, u.URL, maxResponseBodyBytes); err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve the latest Todoist backup")
	}

	if err := gdrive.UploadFile(
		&buf,
		u.Version,
		c.GoogleDrive,
	); err != nil {
		log.Fatal().Err(err).Msg("Unable to upload a file to Google Drive")
	}
}

func main() {
	g := make(chan os.Signal, 1)
	signal.Notify(g, os.Interrupt)

	oneshot := flag.Bool("oneshot", false, "whether to run one backup and exit")
	cf := flag.String("config", "", "the path to a configuration file")
	flag.Parse()

	if *cf == "" {
		fmt.Print(help)
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

	if err := c.GoogleDrive.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid Google Drive config")
	}

	dur, err := time.ParseDuration(c.General.BackupInterval)

	if err != nil {
		log.Fatal().Err(err).Msg("Could not parse the backup interval")
	}

	// Run the first backup right away so we can identify issues
	log.Info().Msg("running initial backup")
	runBackup(c)

	if *oneshot {
		log.Info().Msg("oneshot selected, exiting")
		os.Exit(0)
	}

	k := time.NewTicker(dur)
	for {
		select {
		case <-k.C:
			log.Info().Msg("running periodic backup")
			runBackup(c)
		case <-g:
			log.Info().Msg("Received interrupt. Stopping.")
			os.Exit(0)
		}
	}
}
