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

package mailer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/github"
	"github.com/GoogleCloudPlatform/issuetracker/pkg/internal/testutil"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

// TestMailer is an integration test for checking that Mailer is working correctly
func TestMailer(t *testing.T) {
	instOptions, err := testutil.GetOptions()
	if err != nil {
		t.Errorf("Environment Error: %v", err)
	}

	inst, err := aetest.NewInstance(&instOptions)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()
	// Create a fake user
	user := github.User{
		ID:      1234,
		FireKey: "fakeID", // Fake firebase ID
		Login:   "test",
		Email:   "test@foo.com",
	}
	if err := user.Add(); err != nil {
		t.Errorf("User: Failed to add User to DB: %v", err)
	}
	err = user.Subscribe(
		"GoogleCloudPlatform/nodejs-docs-samples",
		github.EmailPreference{
			IssueOpen:   github.Daily,
			IssueClose:  github.Never,
			IssueReopen: github.Weekly,
			NewComment:  github.Weekly,
			NoComment:   github.Monthly,
		},
	)
	err = user.Subscribe(
		"GoogleCloudPlatform/java-docs-samples",
		github.EmailPreference{
			IssueOpen:   github.Daily,
			IssueClose:  github.Never,
			IssueReopen: github.Weekly,
			NewComment:  github.Weekly,
			NoComment:   github.Monthly,
		},
		"test@bar.com",
	)
	if err != nil {
		t.Errorf("%v", err)
	}
	urlString := "/emailtask?type=weekly&user=" + user.Login
	req, err := inst.NewRequest("GET", urlString, nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	// Response Recorder to check returned status code
	rr := httptest.NewRecorder()

	// Tests func Mailer(w http.ResponseWriter, r *http.Request)
	handler := http.HandlerFunc(EmailTaskHandler)

	//req = req.WithContext(ctx)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	user.Remove()
}

//TestSendMail tests the app engine Mail API integration
func TestSendMail(t *testing.T) {

	instOptions := aetest.Options{
		AppID: "google.com:github-devrel-issues",
	}

	inst, err := aetest.NewInstance(&instOptions)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/daily", nil)
	ctx := appengine.NewContext(req)

	err = sendMail(ctx, "testing@example.com", "TestEmailSubject", "TestMailBody")
	if err != nil {
		t.Errorf("sendMail() failed with error: %v", err)
	}
}
