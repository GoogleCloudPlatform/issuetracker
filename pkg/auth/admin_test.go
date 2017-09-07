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

package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/internal/testutil"

	"github.com/gorilla/mux"

	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/user"
)

// TestServeHttpWithoutLogin ensures that routes are not accessible to anonymous users
func TestServeHTTPWithoutLogin(t *testing.T) {

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
	//setup a dummy handler
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// Response Recorder to check returned status code
	rr := httptest.NewRecorder()

	handler := RequireAdmin{r}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("handler returned wrong status code when Not Logged In: got %v want %v",
			status, http.StatusFound)
	}
}

// TestServeHttpWithLogin ensures that routes are only accessible to admins
func TestServeHTTPWithLogin(t *testing.T) {

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
	// Test the request coming from a user who is logged in
	aetest.Login(&user.User{Email: "test@foo.com", Admin: false}, req)
	//setup a dummy handler
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// Response Recorder to check returned status code
	rr := httptest.NewRecorder()

	handler := RequireAdmin{r}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("handler returned wrong status code when Logged In: got %v want %v",
			status, http.StatusFound)
	}
}
