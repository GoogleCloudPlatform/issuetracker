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
	"strconv"
	"testing"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/internal/testutil"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

// TestUserCRUD tests CRUD operations on the database for User objects
func TestUserCRUD(t *testing.T) {
	if DB.HasTable(&User{}) == false {
		t.Errorf("User: CRUD operations failed, table not found")
	} else {
		x := uint64(1)
		// Test Adds & Updates
		user := &User{
			ID:      123 + x,
			FireKey: "fake" + strconv.FormatUint(x, 10), // Fake firebase ID
			Login:   "test" + strconv.FormatUint(x, 10),
			Email:   "test" + strconv.FormatUint(x, 10) + "@foo.com",
		}
		if err := user.Add(); err != nil {
			t.Errorf("User: Failed to add User to DB: %v", err)
		}
		u, err := FindUserByLogin(user.Login)
		if err != nil {
			t.Errorf("User: Failed to retreive user: %v", err)
		}
		u.Email = "updated-" + user.Email
		if err := u.Update(); err != nil {
			t.Errorf("User: Failed to Update user: %v", err)
		}

		// Test GetUsers
		if _, err := GetUsers(); err != nil {
			t.Errorf("Error creating users, got %v", err)
		}
		// test deletes
		u, err = FindUserByLogin(user.Login)
		if err != nil {
			t.Errorf("User: Failed to Delete user: %v", err)
		}
		if err = u.Remove(); err != nil {
			t.Errorf("User: Failed to Delete user: %v", err)
		}
	}
}

//TestSubscriptions tests CRUD operations on the database for Subscriptions
func TestSubsciptions(t *testing.T) {

	user := &User{
		ID:      1234,
		FireKey: "fakeID", // Fake firebase ID
		Login:   "test",
		Email:   "test@foo.com",
	}
	if err := user.Add(); err != nil {
		t.Errorf("User: Failed to add User to DB: %v", err)
	}
	for x := 0; x < 10; x++ {
		err := user.Subscribe(
			"GoogleCloudPlatform/nodejs-docs-samples-"+strconv.Itoa(x),
			EmailPreference{
				IssueOpen:   Daily,
				IssueClose:  Never,
				IssueReopen: Weekly,
				NewComment:  Weekly,
				NoComment:   Monthly,
			},
		)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	err := user.UpdateSubscription(
		"GoogleCloudPlatform/nodejs-docs-samples-3",
		&Subscription{
			DefaultEmail: "default@go.co",
			EmailPreference: EmailPreference{
				IssueOpen:   Monthly,
				IssueClose:  Monthly,
				IssueReopen: Monthly,
				NoComment:   Never,
				NewComment:  Never,
			},
		},
	)
	if err != nil {
		t.Errorf("Update Subscription: %v", err)
	}
	subs, err := user.GetSubscriptions("GoogleCloudPlatform/nodejs-docs-samples-3")
	if err != nil || subs[0].EmailPreference.NewComment != Never {
		t.Error("Failed to get updated subscription")
	}
	if err := user.UnsubscribeAll(); err != nil {
		t.Errorf("%v", err)
	}
	if err := user.Remove(); err != nil {
		t.Errorf("User: Failed to delete user after getting subscriptions: %v", err)
	}
}

//TestRepoFunctions tests functionality of fetching repo data from Github
func TestUpdateRepo(t *testing.T) {
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
	if err = UpdateRepo(ctx, "GoogleCloudPlatform/java-docs-samples"); err != nil {
		t.Errorf("Failed to update repo: %v", err)
	}
}
