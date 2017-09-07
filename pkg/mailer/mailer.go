// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package mailer implements the appengine services required for sending emails
by using BigQuery and Cloud Datastore for processing data from the Github Dataset
*/
package mailer

import (
	"bytes"
	"html/template"
	"net/http"
	"time"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/github"
	"github.com/GoogleCloudPlatform/issuetracker/pkg/templates"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
	"google.golang.org/appengine/taskqueue"
)

// EmailCronHandler handles creation of task queues for each user for that type of email
func EmailCronHandler(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)
	emailType := r.URL.Query().Get("email")
	users, err := github.GetUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, user := range users {
		// Push a task for the user's daily email
		hostHeader := http.Header{}
		hostHeader.Set("Host", "mailer")
		t := taskqueue.Task{
			Header: hostHeader,
			Path:   "/emailtask?type=" + emailType + "&user=" + user.Login,
			Method: "GET",
		}
		_, err := taskqueue.Add(ctx, &t, emailType)
		if err != nil {
			log.Errorf(ctx, "Failed to create email task: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)

}

// EmailTaskHandler handles sending daily emails triggered by a cron job,
// data for the email is pulled from BigQuery
func EmailTaskHandler(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	userLogin := r.URL.Query().Get("user")
	emailType := r.URL.Query().Get("type")
	user, err := github.FindUserByLogin(userLogin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	emailFrequency := getFrequency(emailType)
	// Fetch Email Data for user
	results, err := github.FetchData(ctx, user.Subscriptions, emailFrequency)
	if err != nil {
		log.Errorf(ctx, "Error getting data:%v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, data := range results {

		if isEmpty(ctx, data.Content) {
			log.Infof(ctx, "No %s content for: %s", emailType, data.Email)
			continue
		}
		// Compose email content
		emailContent, err := composeEmailContent(user.Login, emailType, data.Content)
		if err != nil {
			log.Errorf(ctx, err.Error())
		} else {
			// Send out daily email report &  record the notification data
			subject := "GitHub Activity Digest for " + time.Now().Format("Jan 02,2006")
			err = sendMail(ctx, data.Email, subject, emailContent)
			if err != nil {
				log.Errorf(ctx, err.Error())
			} else {
				// Save notification data
				user.AddNotification(ctx, data.Email, emailFrequency, data.Content)
			}
		}
	}
	w.WriteHeader(http.StatusOK)

}

// composeEmailContent populates an email template with issues
func composeEmailContent(user string, emailType string, data []github.Payload) (string, error) {

	pageTemplate, err := template.ParseFiles(templates.Path() + "email.html")
	if err != nil {
		return "", err
	}

	payload := struct {
		User  string
		Repos []github.Payload
		Type  string
	}{
		user,
		data,
		emailType,
	}
	var emailContent bytes.Buffer

	if err := pageTemplate.Execute(&emailContent, payload); err != nil {
		return "", err
	}

	return emailContent.String(), nil

}

// isEmpty checks if the payload to send an email is empty
func isEmpty(ctx context.Context, data []github.Payload) bool {
	sum := 0
	for _, item := range data {
		log.Infof(ctx, "No content on:%v", item)
		itemSum := len(item.OpenIssues) + len(item.ClosedIssues) + len(item.Comments)
		if itemSum != 0 || item.NoComment {
			sum = sum + 1
		}
	}
	if sum == 0 {
		return true
	}
	return false
}

// sendMail sends out an email to the receiver with the given content
func sendMail(ctx context.Context, to string, subject string, body string) error {

	// TODO: Replace this email with one that is configured with App Engine's Mail API
	msg := &mail.Message{
		Sender:   "email@example.com",
		To:       []string{to},
		Subject:  subject,
		HTMLBody: body,
	}
	if err := mail.Send(ctx, msg); err != nil {
		log.Errorf(ctx, "Couldn't send email: %v", err)
		return err
	}

	return nil
}

func getFrequency(email string) github.Frequency {
	switch email {
	case "daily":
		return github.Daily
	case "weekly":
		return github.Weekly
	case "monthly":
		return github.Monthly
	}
	return github.Daily
}
