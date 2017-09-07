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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"

	"google.golang.org/appengine/log"

	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

const endpoint = "https://api.github.com/"

type clientSecrets struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// API makes an authenticated request to the GitHub API using Oauth2 Client ID & Secret
func API(ctx context.Context, path string, params ...string) (*http.Response, error) {

	credentials := parseCredentials(ctx)
	webClient := urlfetch.Client(ctx)
	additional := ""
	for _, s := range params {
		additional = additional + "&" + s
	}
	url := endpoint + path + credentials + additional
	log.Debugf(ctx, "Github API Fetch: %s", endpoint+path+"?"+additional)
	return webClient.Get(url)
}

// parseCredentials parses the json with github api credentials and returns
// the suffix to be added to an api endpoint for authentication
func parseCredentials(ctx context.Context) string {
	clientSecretPath := credentialsPath()
	f, err := ioutil.ReadFile(clientSecretPath)
	if err != nil {
		log.Errorf(ctx, "api.json file cannot be opened: %s %v", clientSecretPath, err)
		return ""
	}
	var secrets clientSecrets
	err = json.Unmarshal(f, &secrets)
	if err != nil {
		log.Errorf(ctx, "Failed to get client ID & Secret: %v", err)
		return ""
	}
	return "?client_id=" + secrets.ClientID + "&client_secret=" + secrets.ClientSecret
}

// credentialsPath returns a canonical path to the credentials.json file
func credentialsPath() string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	basepath = filepath.Join(basepath, "..", "..")
	return basepath + "/pkg/github/api.json"
}
