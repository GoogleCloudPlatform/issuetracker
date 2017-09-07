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

package bq

import (
	"testing"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/internal/testutil"

	"cloud.google.com/go/bigquery"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

// TestFetch tests queryFetcher using a testing table on BigQuery
func TestFetch(t *testing.T) {

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
	if err != nil {
		t.Errorf("TestFetch failed with error: %v", err)
	}
	tq := Select(Columns{{"type", ""}}).From("ghissues.query_test")
	results, err := Fetch(ctx, tq)
	if err != nil {
		t.Errorf("TestFetch failed with error: %v", err)
	}

	// Check results of query
	var row []bigquery.Value
	results.Next(&row)
	want := "IssuesEvent"
	got := row[0].(string)
	if got != want {
		t.Errorf("TestFetch failed got: %v, want: %v", got, want)
	}
}
