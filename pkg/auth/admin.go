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

// Package auth implements helpers for securing routes
package auth

import (
	"html"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
)

// RequireAdmin is an http.Handler which wraps another handler, requiring
// users to be admins before accessing this handler.
type RequireAdmin struct {
	H http.Handler
}

func (ra RequireAdmin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if user.Current(ctx) != nil && user.IsAdmin(ctx) {
		ra.H.ServeHTTP(w, r)
		return
	}

	url := loginURL(ctx, r.URL.String())
	http.Redirect(w, r, url, http.StatusFound)
}

const (
	// Holds the prefix & suffixes to add to project ID for generating login url
	domainPrefix        = "https://"
	appspotDomainSuffix = ".appspot.com/"
	loginRouteSuffix    = "login"
)

// Returns the URL to login using
func loginURL(ctx context.Context, dest string) string {
	url := domainPrefix + appengine.AppID(ctx) + appspotDomainSuffix + loginRouteSuffix
	url = "/login"
	if len(dest) > 0 {
		url = url + "?redirect=" + html.EscapeString(dest)
	}
	return url
}
