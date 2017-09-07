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
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/github/bq"
	"github.com/GoogleCloudPlatform/issuetracker/pkg/internal/testutil"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

// TestFetchers tests IssueFetcher & CommentFetcher operations
func TestFetchers(t *testing.T) {

	instOptions, err := testutil.GetOptions()
	if err != nil {
		t.Errorf("Environment Error: %v", err)
	}

	inst, err := aetest.NewInstance(&instOptions)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/", nil)
	ctx := appengine.NewContext(req)
	m := make(eventOptions)
	m["opened"] = Options{
		Tables:       []string{"ghissues.test_bed"},
		Repositories: []string{"GoogleCloudPlatform/google-cloud-node"},
		Conditions: []string{
			bq.In("type", "IssuesEvent"),
			bq.In("repo_name", "GoogleCloudPlatform/google-cloud-node"),
			bq.In(bq.JExtract("payload", "action"), "opened"),
		},
		Limit: 1,
	}
	issues, err := fetchIssues(ctx, m, "opened")
	if err != nil || len(issues) != 1 {
		t.Errorf("IssueFetcher() failed got %v with error: %v", issues, err)
	}
	o := Options{
		Tables:       []string{"ghissues.test_bed"},
		Repositories: []string{"GoogleCloudPlatform/google-cloud-node"},
		Conditions: []string{
			bq.In("type", "IssueCommentEvent"),
			bq.In("repo_name", "GoogleCloudPlatform/google-cloud-node"),
			bq.In(bq.JExtract("payload", "action"), "created"),
		},
		Limit: 1,
	}
	got, err := fetchComments(ctx, o)
	if err != nil {
		t.Errorf("CommentFetcher() failed with error: %v", err)
	}
	want := int64(319215574)
	if got[0].ID != want {
		t.Errorf(`CommentFetcher() gave invalid result.Expected Comment With ID: 319215574\n
				  Got: %v`, got)
	}
}

func TestAddReopen(t *testing.T) {
	reopen := []Issue{}
	closed := []Issue{}
	after := time.Now()
	before := after.AddDate(0, 0, -1)
	reopen = append(reopen, fakeIssue(123, after))
	closed = append(closed, fakeIssue(123, before))
	if i := validateReopen(reopen, closed); len(i) != 1 {
		t.Errorf("validateReopen() failed got: %v, wanted 1 issue", len(i))
	}
	reopen = append(reopen, fakeIssue(124, before))
	closed = append(closed, fakeIssue(124, after))
	if i := validateReopen(reopen, closed); len(i) != 1 {
		t.Errorf("validateReopen() error got: %v, wanted 1 issue", len(i))
	}
}

func fakeIssue(i int64, t time.Time) Issue {
	return Issue{
		ID:      i,
		Created: t,
	}
}

func TestFetchData(t *testing.T) {
	instOptions, _ := testutil.GetOptions()

	inst, err := aetest.NewInstance(&instOptions)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/daily", nil)
	ctx := appengine.NewContext(req)
	subs := []Subscription{
		Subscription{
			DefaultEmail: "default@go.co",
			Repo:         "GoogleCloudPlatform/java-docs-samples",
			EmailPreference: EmailPreference{
				IssueOpen:   Weekly,
				IssueClose:  Weekly,
				IssueReopen: Never,
				NoComment:   Never,
				NewComment:  Weekly,
			},
		},
	}
	_, err = FetchData(ctx, subs, Weekly)
	if err != nil {
		t.Errorf("Errors while Fetching Data: %v", err)
	}
}
