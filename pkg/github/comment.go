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
	"encoding/json"
	"io/ioutil"
	"github.com/GoogleCloudPlatform/issuetracker/pkg/github/bq"
	"strconv"
	"strings"
	"time"

	"google.golang.org/appengine/log"

	"golang.org/x/net/context"

	"cloud.google.com/go/bigquery"

	"google.golang.org/api/iterator"
)

/*
Comment holds metadata for comments on GitHub Issues
*/
type Comment struct {
	ID        int64     `gorm:"primary_key"` // Github's unique ID for comments created
	IssueID   string    // Parent Issue's ID for the comment
	Body      string    // Comment's body
	Author    string    // author's github login name
	Created   time.Time // timestamp with date of creation
	UpdatedAt time.Time // timestamp with last update
	Repo      string    // API url for the Comment's parent repo
	URL       string    // https url for the comment on github.com
}

// CommentFetcher uses information stored to query the githubarchive dataset for comments
type CommentFetcher struct {
	query bq.SelectBuilder
	Opts  Options
}

// init sets up the initial query Builder for the  IssueFetcher
func (f *CommentFetcher) init() {

	f.query = bq.Select(bq.Columns{
		{"comment.id", "id"},
		{"issue.repository_url", "repo"},
		{"comment.body", "body"},
		{"comment.user.login", "author"},
		{"comment.created_at", "created"},
		{"comment.html_url", "url"},
	}, "payload").
		From(f.Opts.getTables()...).
		And(f.extractConditions()...).
		OrderBy(f.Opts.getOrder()...).
		Limit(f.Opts.getLimits())

}

// extractConditions returns set conditions from f.Opts or the default set of conditions
// for CommentFetcher to use
func (f *CommentFetcher) extractConditions() []string {
	// return set conditions if present
	if len(f.Opts.Conditions) != 0 {
		return f.Opts.Conditions
	}
	// return default conditions in other cases
	conditions := []string{}

	// Add default condition for IssuesEvents
	conditions = append(conditions, bq.In("type", "IssueCommentEvent"))

	// add condition for repositories to query from
	if len(f.Opts.Repositories) != 0 {
		conditions = append(conditions, bq.In("repo.name", f.Opts.Repositories...))
	}

	// Select only comments where are not deleted by default
	if len(f.Opts.Kind) == 0 {
		f.Opts.Kind = append(f.Opts.Kind, "created")
		conditions = append(conditions, bq.In(bq.JExtract("payload", "issue.state"), "open"))
	}
	conditions = append(conditions, bq.In(bq.JExtract("payload", "action"), f.Opts.Kind...))
	return conditions
}

// Fetch uses the fields stored in CommentFetcher and runs a query job on BigQuery
func (f *CommentFetcher) Fetch(ctx context.Context) ([]Comment, error) {
	f.init()
	results, err := bq.Fetch(ctx, f.query)
	if err != nil {
		return nil, err
	}

	var comments []Comment

	for {
		var m map[string]bigquery.Value
		err := results.Next(&m)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		comments = append(comments, mapToComment(m))
	}
	return comments, nil
}

// mapToComment converts a map of results from BigQuery to issue Objects
func mapToComment(m map[string]bigquery.Value) Comment {
	id, err := strconv.ParseInt(m["id"].(string), 10, 64)
	if err != nil {
		id = -1
	}
	repo := m["repo"].(string)
	body := m["body"].(string)
	author := m["author"].(string)
	created, err := time.Parse(time.RFC3339, m["created"].(string))
	url := m["url"].(string)
	body = trimBody(body)
	return Comment{
		ID:      id,
		IssueID: issueIDFromURL(url),
		Repo:    repo,
		Body:    body,
		Author:  author,
		Created: created,
		URL:     url,
	}
}

// Checks against Github's API to see if there were no comments on this repo for the given
// frequency. Calls  api.github.com/:repo/issues/comments with a generated parameter
// of the form ?since=YYYY-MM-DDT00:00:10Z based on email Frequency passed
func checkNoComment(ctx context.Context, repo string, f Frequency) (bool, error) {

	since := getCommentCheckTime(f).Format("2006-01-02")
	since = since + "T00:00:10Z"
	url := "repos/" + repo + "/issues/comments"
	// Get comments from github
	resp, err := API(ctx, url, "since="+since)
	if err != nil {
		log.Errorf(ctx, "Error checking comments for repo %s: %v", repo, err)
		return false, err
	}
	resBody, _ := ioutil.ReadAll(resp.Body)
	var jsonResponse []map[string]interface{}
	json.Unmarshal(resBody, &jsonResponse)
	if len(jsonResponse) == 0 {
		return true, nil
	}
	return false, nil
}

func getCommentCheckTime(f Frequency) (since time.Time) {

	usLoc, _ := time.LoadLocation("America/Los_Angeles")
	today := time.Now().In(usLoc)
	if f == Weekly {
		weekStart := today.AddDate(0, 0, -7)
		since = weekStart
	} else if f == Monthly {
		since = today.AddDate(0, -1, 0)
	} else {
		since = today.AddDate(0, 0, -1)
	}
	return
}

func trimBody(body string) string {
	if len(body) < 120 {
		return body
	}
	return body[0:100] + "..."
}

func issueIDFromURL(url string) string {
	path := strings.SplitAfterN(url, "/", 7)
	path = strings.SplitAfterN(path[len(path)-1], "#", 2)
	result := path[0]
	return result[:len(result)-1]
}
