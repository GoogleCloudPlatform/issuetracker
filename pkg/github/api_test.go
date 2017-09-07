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

	"github.com/GoogleCloudPlatform/issuetracker/pkg/internal/testutil"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

// TestAPI tests authenticated requests to github
func TestAPI(t *testing.T) {
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
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	ctx := appengine.NewContext(req)

	resp, err := API(ctx, "/rate_limit")
	if err != nil {
		t.Error(err)
	}
	if limit := resp.Header.Get("X-RateLimit-Remaining"); limit != "0" {
		t.Errorf("Rate limit exceeded, Remaining: %v", resp.Header)
	}
}
