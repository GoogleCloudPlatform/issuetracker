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

// Package github provides definitions and methods for structs used for querying
// data from the githubarchive dataset
package github

import (
	"testing"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/github/bq"
	"github.com/GoogleCloudPlatform/issuetracker/pkg/internal/testutil"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

// TestCommentFetcher runs a unit test on the BigQuery query to check for errors raised
func TestCommentFetcher(t *testing.T) {

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

	fetcher := CommentFetcher{
		Opts: Options{
			Tables: []string{"ghissues.test_bed"},
			Conditions: []string{
				bq.In("type", "IssueCommentEvent"),
				bq.In("repo_name", "GoogleCloudPlatform/google-cloud-node"),
				bq.In(bq.JExtract("payload", "action"), "created"),
			},
			Limit: 1,
		},
	}
	got, err := fetcher.Fetch(ctx)
	if err != nil {
		t.Errorf("CommentFetcher() failed with error: %v", err)
	}
	want := int64(319215574)
	if got[0].ID != want {
		t.Errorf(`CommentFetcher() gave invalid result.Expected Comment With ID: 319215574\n
				  Got: %v`, got)
	}

}

//Tests checkNoComment
func TestCheckNoComment(t *testing.T) {

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

	if val, err := checkNoComment(ctx, "arjun-rao/go-ae-starter", Daily); err != nil {
		t.Errorf("checkNoComment() returned %v and failed with error: %v", val, err)
	}

}
