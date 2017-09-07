# GitHub Issue Tracker

Github Issue Tracker is a web application hosted on Google Cloud Platform that provides consolidated
email digests of issues on GitHub repositories. Users can subscribe to email notifications
per repository. BigQuery has a [dataset](https://githubarchive.org) that contains
GitHub events over time (githubarchive.org) which is updated hourly. This is used to provide
consolidated email digests of GitHub Issues.

The frequency of email digests is configured by the User and stored in Cloud SQL.
Based on a user's preferences, a consolidated daily/weekly/monthly email is sent out to the user
that summarizes the latest activity on GitHub repositories.
The tool is built using Go, running on app engine as different services.

The entire tool consists of 3 services on app engine: The frontend service that is responsible for
the UI and dealing with operations such as authentication of users;
The backend service handles operations such as monitoring the BigQuery dataset,
updating Cloud SQL tables. The mailer service handles the daily, weekly and monthly cron jobs
 that trigger email notifications to be sent. It is also responsible for composing the
 content for email notifications and sending them out using the app engine Mail API.

The tool demonstrates integration of various App Engine features along with other Google Services
such as BigQuery, Firebase, and Cloud SQL, and building an app using a
[Microservices Architecture][7].

## Disclaimer

This is not an official Google product.

## Quickstart

### Installing dependencies:

See README_dep.md for notes regarding dependencies and configuration

### Creating Test Tables in BigQuery:
Create a BigQuery Table for testing purposes using the following query:

	SELECT * FROM [githubarchive.month.201707] WHERE type IN ('IssuesEvent','IssueCommentEvent)

### Running Locally:
You can then run the app locally:

	cd PROJECT_DIR
	cd services
	dev_appserver.py -A [ProjectID] mailer/app.yaml backend/app.yaml frontend/app.yaml  dispatch.yaml  --show_mail_body=yes

Visit localhost:8080 to view the application


### Deploying to App Engine

	gcloud app deploy mailer/app.yaml backend/app.yaml frontend/app.yaml mailer/queue.yaml mailer/cron.yaml dispatch.yaml

## Products
- [Google BigQuery][2]
- [Google App Engine][4] (Standard Environment)
- [Firebase Authentication using Github][5]
- [Google Cloud SQL][6]

## Language
- [Go][3]

[2]: https://cloud.google.com/bigquery
[3]: https://golang.org
[4]: https://cloud.google.com/appengine/
[5]: https://firebase.google.com/docs/auth/web/github-auth
[6]: https://cloud.google.com/appengine/docs/standard/go/cloud-sql/
[7]: https://cloud.google.com/appengine/docs/standard/go/microservices-on-app-engine