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

// Package github provides definitions and methods for structs used for handling
// Github data from BigQuery and Cloud SQL. Methods for making authenticated requests to
// the GitHub API are implemented in this package
package github

import (
	"strconv"
	"time"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/github/bq"

	"golang.org/x/net/context"

	"cloud.google.com/go/bigquery"

	"google.golang.org/api/iterator"
)

/*
Issue holds metadata for GitHub Issues
*/
type Issue struct {
	ID        int64     `gorm:"primary_key"` // Github's unique ID for issues created
	Number    int       // issue number that is specific to a repository
	Title     string    // title for the issue
	Author    string    // author's github login name
	Created   time.Time // timestamp with date of creation
	UpdatedAt time.Time // timestamp with last update
	Repo      string    // API url for the issue's parent repo
	URL       string    // https url for the issue on github.com
}

// IssueFetcher uses information stored to query the githubarchive dataset for issues
type IssueFetcher struct {
	query bq.SelectBuilder
	Opts  Options
}

// init sets up the default query for the IssueFetcher
func (f *IssueFetcher) init() {

	f.query = bq.Select(bq.Columns{
		{"issue.id", "id"},
		{"issue.number", "number"},
		{"issue.title", "title"},
		{"issue.user.login", "author"},
		{"issue.updated_at", "created"},
		{"issue.repository_url", "repo"},
		{"issue.html_url", "url"},
	}, "payload").
		From(f.Opts.getTables()...).
		And(f.extractConditions()...).
		OrderBy(f.Opts.getOrder()...).
		Limit(f.Opts.getLimits())
}

// extractConditions returns set conditions from f.Opts or the default set of conditions
// for IssueFetcher to use
func (f *IssueFetcher) extractConditions() []string {
	// return set conditions if present
	if len(f.Opts.Conditions) != 0 {
		return f.Opts.Conditions
	}
	// return default conditions in other cases
	conditions := []string{}

	// Add default condition for IssuesEvents
	conditions = append(conditions, bq.In("type", "IssuesEvent"))

	// add condition for repositories to query from
	if len(f.Opts.Repositories) != 0 {
		conditions = append(conditions, bq.In("repo.name", f.Opts.Repositories...))
	}

	// Select only IssueEvents which were opened by default
	if len(f.Opts.Kind) == 0 {
		f.Opts.Kind = append(f.Opts.Kind, "opened")
	}
	conditions = append(conditions, bq.In(bq.JExtract("payload", "action"), f.Opts.Kind...))
	return conditions
}

// Fetch uses the data stored in IssueFetcher and runs a query job on BigQuery
func (f *IssueFetcher) Fetch(ctx context.Context) ([]Issue, error) {
	f.init()
	results, err := bq.Fetch(ctx, f.query)
	if err != nil {
		return nil, err
	}

	var issues []Issue

	for {
		var m map[string]bigquery.Value
		err := results.Next(&m)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		issues = append(issues, mapToIssue(m))
	}
	return issues, nil
}

// mapToIssue converts a map of results from BigQuery to issue Objects
func mapToIssue(m map[string]bigquery.Value) Issue {
	id, err := strconv.ParseInt(m["id"].(string), 10, 64)
	if err != nil {
		id = -1
	}
	number, err := strconv.Atoi(m["number"].(string))
	if err != nil {
		number = -1
	}
	repo := m["repo"].(string)
	title := m["title"].(string)
	author := m["author"].(string)
	created, err := time.Parse(time.RFC3339, m["created"].(string))
	url := m["url"].(string)

	return Issue{
		ID:      id,
		Number:  number,
		Title:   title,
		Author:  author,
		Created: created,
		Repo:    repo,
		URL:     url,
	}
}
