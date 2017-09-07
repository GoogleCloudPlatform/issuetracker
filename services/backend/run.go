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
	"net/http"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/auth"
	"github.com/GoogleCloudPlatform/issuetracker/pkg/backend"

	"github.com/gorilla/mux"
)

func init() {
	registerHandlers()
}

// registerHandlers defines the routes in the application
func registerHandlers() {
	r := mux.NewRouter()

	// Endpoint for debugging - requires admin access
	r.Methods("GET").Path("/debug/{id}").Handler(backend.GetHandler(backend.UserGet))

	api := r.PathPrefix("/api/").Subrouter()

	// Auth API
	api.HandleFunc("/auth", backend.AuthTokenHandler).Methods("GET")

	// Subscription API
	api.Methods("GET").Path("/subscriptions").Handler(backend.GetHandler(backend.GetSubs))
	api.Methods("POST").Path("/subscriptions/add").Handler(backend.GetHandler(backend.AddSubs))
	api.Methods("POST").Path("/subscriptions/update").Handler(backend.GetHandler(backend.UpdateSub))
	api.Methods("POST").Path("/subscriptions/remove").Handler(backend.GetHandler(backend.DelSubs))

	// User API
	api.Methods("GET").Path("/users/repos").Handler(backend.GetHandler(backend.GetRepos))
	api.Methods("POST").Path("/users/add").Handler(backend.GetHandler(backend.UserAdd))
	api.Methods("POST").Path("/users/update").Handler(backend.GetHandler(backend.UserUpdate))

	// Notifications API
	api.Methods("GET").Path("/notifications").Handler(backend.GetHandler(backend.GetNotifications))

	//Resond to App Engine health checks
	r.Methods("GET").Path("/_ah/health").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

	// Set up 404 Handlers for all miscellenous routes
	api.NotFoundHandler = http.RedirectHandler("/", http.StatusNotFound)
	r.NotFoundHandler = http.RedirectHandler("/", http.StatusNotFound)

	// Route all requests through the Mux and add CSRF protection
	http.Handle("/api/", r)

	// Prevent unauthorised requests to backend server
	http.Handle("/", auth.RequireAdmin{r})

}
