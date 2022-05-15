# ROADMAP

- Ensure that the OneDrive credentials are scoped only to the given directory.

  To do this, request authorization for the "Files.ReadWrite.AppFolder" scope (https://docs.microsoft.com/en-us/onedrive/developer/rest-api/concepts/special-folders-appfolder). The user authorizes the app to access their app folder. App folders are only compatible with personal OneDrive accounts.

  Note that it's only clear how to request authorization scopes from the OneDrive API reference code, so we'll need to see if it's possible to request this authorization while also obtaining a token to use with `net/http` (as the current code does).

  Maybe do some background reading on the libraries involved in the reference authentication code to see what's possible.

- Use a logger instead of print statements since this will be running remotely. Also log whenever a retry is attempted