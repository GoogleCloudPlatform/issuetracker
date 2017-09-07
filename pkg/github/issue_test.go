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

// TestIssueFetcher runs a unit test on the BigQuery query to check for errors raised
func TestIssueFetcher(t *testing.T) {

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

	fetcher := IssueFetcher{
		Opts: Options{
			Tables: []string{"ghissues.test_bed"},
			Conditions: []string{
				bq.In("type", "IssuesEvent"),
				bq.In("repo_name", "GoogleCloudPlatform/google-cloud-node"),
				bq.In(bq.JExtract("payload", "action"), "opened"),
			},
			Limit: 1,
		},
	}
	got, err := fetcher.Fetch(ctx)
	if err != nil {
		t.Errorf("IssueFetcher() failed with error: %v", err)
	}
	want := 2498
	if got[0].Number != want {
		t.Errorf("IssueFetcher() gave invalid result.\nExpected Issue #2460\nGot: %v", got)
	}

}
