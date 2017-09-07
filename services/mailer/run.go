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

	"github.com/GoogleCloudPlatform/issuetracker/pkg/mailer"

	"github.com/gorilla/mux"
)

// Function that defines the routes in the application
func init() {
	r := mux.NewRouter()
	r.HandleFunc("/cron", mailer.EmailCronHandler)
	r.HandleFunc("/emailtask", mailer.EmailTaskHandler)
	r.NotFoundHandler = http.RedirectHandler("/", http.StatusForbidden)
	// Route all requests through the Mux
	http.Handle("/", r)

}
