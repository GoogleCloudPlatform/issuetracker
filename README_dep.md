# Instructions for configuration

## Installing Dependencies

Run `go get` in the following directories:

*  `services/mailer/`
*  `services/backend/`


## Firebase Credentials

Download the service account credentials from Firebase Console and save it as
`pkg/auth/credentials.json`

Replace the Firebase credentials in both files in `/services/frontend/src/app/environment/`
with your Firebase API credentials. Read the
[Documentation](https://firebase.google.com/docs/web/setup) here for more details on obtaining
credentials.

## Cloud SQL Configuration

Edit `/services/mailer/app.yaml` and `/services/backend/app.yaml` and replace the
database credentials with the one that you have setup for your application.
Create a new database called `ghdata` in Cloud SQL and in your local instance of MySQL.

## GitHub Credentials

Get the app's GitHub oAuth Client ID and Client Secret and store it in `pkg/github/api.json`
using the following format:

    {
        "client_id":"CLIENT ID HERE",
        "client_secret":"CLIENT SECRET HERE"
    }

## Sending Emails

While the local development sender doesn't send out emails directly, on App Engine, using the Mail
API, you can send out actual emails.
Edit the email address set on `line 161` in `/pkg/mailer/mailer.go` with your email address, and add
the same address to the `Email API authorized senders` list under App Engine settings in the
Google Cloud Platform Console

## Environment Variables for Local Development

Setup the following Environment variables for testing this application locally:

    "PROJECT_ID": "github-issue-tracker",
    "CLOUDSQL_USER":"root",
    "CLOUDSQL_PASSWORD": "password",
    "MYSQL_HOST":"127.0.0.1",
    "MYSQL_PORT":"3306",
    "RUN_WITH_DEVAPPSERVER":"true"

Replace the values for `CLOUDSQL_USER, CLOUDSQL_PASSWORD` variables with the credentials
you use with your local mysql database.

# Notes about dependencies

This project was initially designed to use [dep](https://github.com/golang/dep) for dependency
management. It turned out that Google App Engine Standard and `dep` have compatibility
issues when using `gcloud`. The following discussion presents the case for not using `dep`
with this project.

## Using dep with gcloud

When vendored dependancies are present, `gcloud app deploy` breaks and deployment fails. See the
discussion [here](https://groups.google.com/forum/#!topic/google-appengine-go/Xooyiq3kFTI)

## Using dep with goapp

Per the discussion mentioned above, if you were to use `goapp` to deploy, the deployment works
without issues. However `goapp` does not seem to be uploading .html files. Deploying the following
[project](https://github.com/arjun-rao/go-ae-starter) to GAE using goapp fails to upload .html files
while it does upload html files when using `gcloud` - however if vendoring via the `dep` tool is
used, it breaks the build.

## Conclusion

We have not used dep for managing deployments in this project. It maybe worth exploring at a future
time when `dep` is more mature and possibly integrated into standard Go tooling.

Dependencies for this project:

* "golang.org/x/net/context"
* "cloud.google.com/go/bigquery"
* "google.golang.org/appengine".
* "google.golang.org/api/iterator"
* "google.golang.org/appengine/mail"
* "github.com/lann/builder"
* "github.com/gorilla/mux"
* "github.com/wuman/firebase-server-sdk-go"
* "github.com/go-sql-driver/mysql"
* "github.com/jinzhu/gorm"

