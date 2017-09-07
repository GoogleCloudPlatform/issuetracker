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

package github

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/appengine/log"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/github/bq"

	"golang.org/x/net/context"
)

type eventOptions map[string]Options
type eventRepo map[string][]string

// Payload is the type that contains email data for one repo
type Payload struct {
	RepoName       string
	OpenIssues     []Issue
	ClosedIssues   []Issue
	Comments       []Comment
	NoComment      bool
	NoCommentSince time.Time
}

// EmailPayload is the type that contains email data for one Email to be sent
type EmailPayload struct {
	Email   string
	Content []Payload
}

var ctx context.Context

// FetchData constructs bigquery requests for fetching appropriate data for email notifications
func FetchData(c context.Context,
	subscriptions []Subscription, emailType Frequency) ([]EmailPayload, error) {

	var errors error
	//setup global context to use for logging
	ctx = c
	eventReposMap := mapMaker(subscriptions, emailType)
	// Get query options for using IssueFetcher and CommentFetcher
	eventOptionsMap := optionMaker(eventReposMap, emailType)
	// Call IssueFetcher for Open,Closed,Reopen []Issues
	openIssues, err := fetchIssues(ctx, eventOptionsMap, "opened")

	if err != nil {
		log.Errorf(ctx, "Error fetching open issues: %v", err)
		errors = fmt.Errorf("Error with Open Issues: %v", err)
	}
	closedIssues, err := fetchIssues(ctx, eventOptionsMap, "closed")
	if err != nil {
		log.Errorf(ctx, "Error fetching closed issues: %v", err)
		errors = fmt.Errorf("%v\nError with Closed Issues: %v", errors, err)
	}
	reopenedIssues, err := fetchIssues(ctx, eventOptionsMap, "reopened")
	if err != nil {
		log.Errorf(ctx, "Error fetching reopened issues: %v", err)
		errors = fmt.Errorf("%v\nError with Reopened Issues: %v", errors, err)
	}

	// Call CommentFetcher for Comments
	comments, err := fetchComments(ctx, eventOptionsMap["comment"])
	if err != nil {
		log.Errorf(ctx, "Error fetching comments issues: %v", err)
		errors = fmt.Errorf("%v\nError with Comments: %v", errors, err)
	}
	// Remove all closed issues from open issues
	openIssues = issueDiff(openIssues, closedIssues)
	// Add all reopen to open if reopen date is greater than closed date for common issues
	reopenedIssues = validateReopen(reopenedIssues, closedIssues)
	for _, issue := range reopenedIssues {
		openIssues = append(openIssues, issue)
	}
	// Get list of repos with no comment
	repoWithNoComment := []string{}
	for _, repo := range eventReposMap["nocomment"] {
		if add, err := checkNoComment(ctx, repo, emailType); add && err == nil {
			repoWithNoComment = append(repoWithNoComment, repo)
		}
	}
	log.Infof(ctx, "No Comments on: %v", repoWithNoComment)
	if len(openIssues)+len(closedIssues)+len(comments)+len(repoWithNoComment) == 0 {
		// No data to send emails
		log.Infof(ctx, "No emails sent for user: %v", subscriptions[0].UserID)
		return nil, nil
	}

	//  Sort all fetched data by repo:
	repoData := make(map[string]Payload)

	for _, issue := range openIssues {
		key := repoFromIssue(issue)
		value := repoData[key]
		value.OpenIssues = append(value.OpenIssues, issue)
		value.RepoName = key
		repoData[key] = value
	}
	for _, issue := range closedIssues {
		key := repoFromIssue(issue)
		value := repoData[key]
		value.ClosedIssues = append(value.ClosedIssues, issue)
		value.RepoName = key
		repoData[key] = value
	}
	for _, comment := range comments {
		key := repoFromComment(comment)
		value := repoData[key]
		value.RepoName = key
		value.Comments = append(value.Comments, comment)
		repoData[key] = value
	}

	for _, repo := range repoWithNoComment {
		value := repoData[repo]
		value.RepoName = repo
		value.NoComment = true
		value.NoCommentSince = getCommentCheckTime(emailType)
		repoData[repo] = value
	}

	//Sort all subscriptions by email
	emailRepoMap := make(map[string][]string)
	for _, sub := range subscriptions {
		emailRepoMap[sub.DefaultEmail] = append(emailRepoMap[sub.DefaultEmail], sub.Repo)
		data := repoData[sub.Repo]
		repoData[sub.Repo] = data
	}

	// Create payload for emails
	results := []EmailPayload{}
	for id, repos := range emailRepoMap {
		payload := EmailPayload{
			Email:   id,
			Content: getPayloads(repoData, repos...),
		}
		results = append(results, payload)
	}
	return results, errors
}

// mapMaker returns a map with repo names sorted according to Github Event types
// It compares the EmailPreferences for a subscription wiht the emailType frequency before adding
func mapMaker(subscriptions []Subscription, emailType Frequency) eventRepo {

	// Map to hold Event Kind and List of Repos to query for each Event Kind
	repos := make(eventRepo)
	for _, sub := range subscriptions {
		if sub.EmailPreference.IssueOpen == emailType {
			repos["opened"] = append(repos["opened"], sub.Repo)
		}
		if sub.EmailPreference.IssueClose == emailType {
			repos["closed"] = append(repos["closed"], sub.Repo)
		}
		if sub.EmailPreference.IssueReopen == emailType {
			repos["reopened"] = append(repos["reopened"], sub.Repo)
		}
		if sub.EmailPreference.NewComment == emailType {
			repos["comment"] = append(repos["comment"], sub.Repo)
		}
		if sub.EmailPreference.NoComment == emailType {
			repos["nocomment"] = append(repos["nocomment"], sub.Repo)
		}
	}
	log.Infof(ctx, "mapMaker: %v", repos)
	return repos
}

func optionMaker(m map[string][]string, emailType Frequency) eventOptions {

	options := make(eventOptions)

	for event, repos := range m {
		if isIssueEvent(event) {
			o := Options{
				Repositories: repos,
				Conditions: []string{
					bq.In("type", "IssuesEvent"),
					bq.In("repo.name", repos...),
					bq.In(bq.JExtract("payload", "action"), event),
				},
			}
			o.SetTables(emailType)
			options[event] = o
		}
		if event == "comment" {
			o := Options{
				Repositories: repos,
				Conditions: []string{
					bq.In("type", "IssueCommentEvent"),
					bq.In("repo.name", repos...),
					bq.In(bq.JExtract("payload", "issue.state"), "open"),
					bq.In(bq.JExtract("payload", "action"), "created"),
				},
			}
			o.SetTables(emailType)
			options[event] = o
		}
	}
	log.Infof(ctx, "optionMaker: %v", options)
	return options
}

func isIssueEvent(e string) bool {
	issueEvent := map[string]bool{
		"opened":   true,
		"closed":   true,
		"reopened": true,
	}
	if issueEvent[e] {
		return true
	}
	return false
}

// Retrieves issues from BigQuery
func fetchIssues(ctx context.Context, m eventOptions, event string) ([]Issue, error) {

	fetcher := IssueFetcher{
		Opts: m[event],
	}
	if len(fetcher.Opts.Repositories) == 0 {
		return []Issue{}, nil
	}
	results, err := fetcher.Fetch(ctx)
	if err != nil {
		log.Errorf(ctx, "Issue Fetch: %v", err)
		return nil, err
	}
	log.Infof(ctx, "fetchIssues: %v", len(results))
	return results, nil
}

// Retrieves issues from BigQuery
func fetchComments(ctx context.Context, o Options) ([]Comment, error) {

	if len(o.Repositories) == 0 {
		return []Comment{}, nil
	}
	fetcher := CommentFetcher{
		Opts: o,
	}
	results, err := fetcher.Fetch(ctx)
	if err != nil {
		log.Errorf(ctx, "Issue Fetch: %v", err)
		return nil, err
	}
	log.Infof(ctx, "fetchComments: %v", len(results))
	return results, nil
}

func repoFromIssue(i Issue) string {
	path := strings.SplitAfterN(i.Repo, "/", 5)
	return path[len(path)-1]
}
func repoFromComment(c Comment) string {
	path := strings.SplitAfterN(c.Repo, "/", 5)
	return path[len(path)-1]
}

// issueDiff is a helper method to remove B from A
func issueDiff(A []Issue, B []Issue) []Issue {
	mb := map[int64]bool{}
	for _, x := range B {
		mb[x.ID] = true
	}
	result := []Issue{}
	for _, x := range A {
		if _, ok := mb[x.ID]; !ok {
			result = append(result, x)
		}
	}
	return result
}

// validateReopen is a helper method to return valid issues that are open
func validateReopen(reopen []Issue, closed []Issue) []Issue {
	mb := map[int64]Issue{}
	for _, x := range closed {
		mb[x.ID] = x
	}
	result := []Issue{}
	for _, x := range reopen {
		if _, ok := mb[x.ID]; !ok || x.Created.After(mb[x.ID].Created) {
			result = append(result, x)
		}
	}
	return result
}

// returns payload for given repos
func getPayloads(m map[string]Payload, repos ...string) []Payload {
	results := []Payload{}
	for _, repo := range repos {
		results = append(results, m[repo])
	}
	return results
}
