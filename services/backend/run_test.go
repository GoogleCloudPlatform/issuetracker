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

package backend

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/github"
	"github.com/GoogleCloudPlatform/issuetracker/services/backend/webtest"
)

var wt *webtest.W

func TestMain(m *testing.M) {
	serv := httptest.NewServer(nil)
	wt = webtest.New(nil, serv.Listener.Addr().String())
	os.Exit(m.Run())
}
func TestInvalidUser(t *testing.T) {
	bodyContains(t, wt, "/api/user/0", "No such user")
}

func TestGetUser(t *testing.T) {
	user := github.User{
		ID:      123,
		Login:   "test",
		Email:   "foo@gmail.com",
		FireKey: "123asfasfd",
	}
	user.Add()
	user.Subscribe("GoogleCloudPlatform/java-docs-samples",
		github.EmailPreference{IssueOpen: github.Daily})
	bodyContains(t, wt, "/api/user/test", "foo@gmail.com")
	user.Remove()
}

func TestUserAdd(t *testing.T) {
	var body bytes.Buffer
	m := multipart.NewWriter(&body)
	m.WriteField("github_login", "test")
	m.WriteField("github_id", "123456")
	m.WriteField("github_email", "foo@baz.com")
	m.WriteField("fb_id", "test123")
	m.CreateFormFile("image", "")
	m.Close()
	postContains(t, wt, "/api/users/add", m, body, "foo@baz.com")
	u, err := github.FindUserByLogin("test")
	if err != nil {
		t.Error("Failed to create user")
	}
	u.Remove()
}

func bodyContains(t *testing.T, wt *webtest.W, path, contains string) (ok bool) {
	body, _, err := wt.GetBody(path)
	if err != nil {
		t.Error(err)
		return false
	}
	if !strings.Contains(body, contains) {
		t.Errorf("got %s, want %s", body, contains)
		return false
	}
	return true
}

func postContains(t *testing.T, wt *webtest.W, path string,
	m *multipart.Writer, body bytes.Buffer, contains string) (ok bool) {
	resp, err := wt.Post(path, "multipart/form-data; boundary="+m.Boundary(), &body)
	if err != nil {
		t.Fatal(err)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	if !strings.Contains(string(respBody), contains) {
		t.Errorf("got %s, want %s", string(respBody), contains)
		return false
	}
	return true
}
