# Todoist Backup Job

This is a program for backing up Todoist data on Google Drive.

## Create a service account

The Todoist backup job authenticates to the Google Drive API via a Google
Cloud service account.

Follow [these instructions](https://developers.google.com/workspace/guides/create-credentials#service-account) to create a service account.

## Run the backup job

There are two ways to run the backup job.

### Execute a binary

Download Go version 1.18+ and run the following command:

```bash
$ go build -o tbackups main.go
```

Add `tbackups` to your `PATH`.


Run the following command:

```bash
$ tbackups -config=config.yaml [-oneshot]
```

You must provide a `-config` flag with the path to a config file, which is explained in the next section.

### Run a container

You can run the Todoist backup job as a container:

```bash
$ CONFIG_PATH=<absolute path to config>;
$ CREDENTIALS_PATH=<absolute path to credentials file>;
$ docker run -v $CONFIG_PATH:/config.yaml -v $CREDENTIALS_PATH:/credentials.json ptgott/todoist-backups:0.10 -config=/config.yaml;
```

Make sure that your config file's `credentials_path` field points to the location of your credentials file _within the container_.

## Config file

The config file must include the following options in YAML format:

```yaml
general:

  # The API key retrieved from Todoist
  todoist_api_key: 123abc

  # How often to conduct the backup. A duration string like "1m", "4h", or "3d"
  backup_interval: 24h

google_drive:

  # Path to a Google Workspace credentials file, which you can export for the
  # service account that you created for this app. This must be a service 
  # account credentials file, rather than an OAuth2.0 token.
  credentials_path: credentials.json

  # Name of the Google Drive directory you want to write backups to
  folder_name: todoist-backups

```

Note that the folder in `folder_name` must be shared with the service account you create
for Todoist backups. The service account's email address will be provided
on creation. The Todoist backup job will be limited to this directory.

You can optionally use the `-oneshot` flag to create a single backup without
running the job as a daemon.