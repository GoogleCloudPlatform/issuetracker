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

/*
Package core implements the core backend service for the app */
package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/auth"
	"github.com/GoogleCloudPlatform/issuetracker/pkg/github"

	"google.golang.org/appengine"

	"github.com/gorilla/mux"
)

// StatusHandler is used for debugging the app as an admin
func StatusHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{'message':'Running'}")
	for k, v := range r.Header {
		fmt.Fprintf(w, "%v = %v\n", k, v)
	}

}

// UserGet retrieves a given user from the request
func UserGet(w http.ResponseWriter, r *http.Request) *AppError {
	login := mux.Vars(r)["id"]
	user, err := github.FindUserByLogin(login)
	if err != nil {
		return appErrorf(err, "No such user: %v", login)
	}
	writeJSON(w, user)
	return nil
}

// GetSubs retrieves subscriptions for a given user
func GetSubs(w http.ResponseWriter, r *http.Request) *AppError {

	user, err := getAuthenticatedUser(w, r)
	if err != nil {
		return appErrorf(err, "No such user: %v", user.Login)
	}
	subs, _ := user.GetSubscriptions()
	writeJSON(w, subs)
	return nil
}

// GetNotifications retrieves notifications sent for a given user
func GetNotifications(w http.ResponseWriter, r *http.Request) *AppError {

	user, err := getAuthenticatedUser(w, r)
	if err != nil {
		return appErrorf(err, "No such user: %v", user.Login)
	}
	notifs, _ := user.GetNotifications(0)
	writeJSON(w, notifs)
	return nil
}

// GetRepos retrieves open issue counts for repositories that a user subscribes to
func GetRepos(w http.ResponseWriter, r *http.Request) *AppError {

	user, err := getAuthenticatedUser(w, r)
	if err != nil {
		return appErrorf(err, "No such user: %v", user.Login)
	}
	repos, _ := user.GetRepos()
	writeJSON(w, repos)
	return nil
}

// AddSubs retrieves subscriptions for a given user
func AddSubs(w http.ResponseWriter, r *http.Request) *AppError {

	ctx := appengine.NewContext(r)
	user, err := getAuthenticatedUser(w, r)
	if err != nil {
		return appErrorf(err, "No such user: %v", user.Login)
	}
	repo := r.FormValue("repo")
	if err = github.UpdateRepo(ctx, repo); err != nil {
		log.Printf("Github API Fetch error: %v", err)
		return appErrorf(err, "Couldn't subscribe to repo: %v", repo)
	}
	if err := user.Subscribe(repo, github.NewPreference()); err != nil {
		writeJSON(w, status{err, "Could not subscribe to repo", 500})
		return appErrorf(err, "Couldn't subscribe to repo: %v", repo)
	}

	subs, _ := user.GetSubscriptions()
	writeJSON(w, subs)
	return nil
}

// UpdateSub updates subscriptions for a given user
func UpdateSub(w http.ResponseWriter, r *http.Request) *AppError {

	user, err := getAuthenticatedUser(w, r)
	if err != nil {
		return appErrorf(err, "No such user: %v", user.Login)
	}
	repo := r.FormValue("repo")
	defaultEmail := r.FormValue("defaultEmail")
	settings := []byte(r.FormValue("settings"))
	preferences := github.EmailPreference{}
	err = json.Unmarshal(settings, &preferences)
	if err != nil {
		return appErrorf(err, "Couldn't get settings for repo: %v", repo)
	}
	subs, _ := user.GetSubscriptions(repo)
	if len(subs) == 1 {
		sub := subs[0]
		sub.EmailPreference = preferences
		if len(defaultEmail) != 0 {
			sub.DefaultEmail = defaultEmail
		}
		if err := user.UpdateSubscription(repo, &sub); err == nil {
			writeJSON(w, sub)
			return nil
		}
	}

	return appErrorf(err, "Couldn't update settings for repo: %v", repo)

}

// DelSubs retrieves subscriptions for a given user
func DelSubs(w http.ResponseWriter, r *http.Request) *AppError {

	user, err := getAuthenticatedUser(w, r)
	if err != nil {
		return appErrorf(err, "No such user: %v", user.Login)
	}
	repo := r.FormValue("repo")
	if err := user.Unsubscribe(repo); err != nil {
		writeJSON(w, status{err, "Could not unsubscribe repo", 500})
		return appErrorf(err, "Couldn't unsubscribe repo: %v", repo)
	}
	writeJSON(w, status{err, "ok", 200})
	return nil
}

// UserAdd handles creation of a new user
func UserAdd(w http.ResponseWriter, r *http.Request) *AppError {
	var user github.User
	user.Login = r.FormValue("github_login")
	user.FireKey = r.FormValue("fb_id")
	user.ID, _ = strconv.ParseUint(r.FormValue("github_id"), 10, 64)
	user.Email = r.FormValue("github_email")
	w.Header().Set("Content-Type", "application/json")
	payload, _ := json.Marshal(user)
	w.Write([]byte(payload))
	err := user.Add()
	if err != nil {
		return appErrorf(err, "Error creating user")
	}
	return nil
}

// UserUpdate updates a user's preferences
func UserUpdate(w http.ResponseWriter, r *http.Request) *AppError {

	user, err := getAuthenticatedUser(w, r)
	if err != nil {
		return appErrorf(err, "No such user: %v", user.Login)
	}
	if email := r.FormValue("email"); len(email) != 0 {
		user.Email = email
	}
	w.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(user)
	w.Write([]byte(response))

	if err := user.Update(); err != nil {
		return appErrorf(err, "Error updating user")
	}
	return nil

}

// GetHandler wraps a handler func with error handling logic
// http://blog.golang.org/error-handling-and-go
type GetHandler func(http.ResponseWriter, *http.Request) *AppError

// AppError xports type for a Handler errors
type AppError struct {
	Error   error
	Message string
	Code    int
}

func (fn GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		log.Printf("Handler error: status code: %d, message: %s, underlying err: %#v",
			e.Code, e.Message, e.Error)

		http.Error(w, e.Message, e.Code)
	}
}

func appErrorf(err error, format string, v ...interface{}) *AppError {
	return &AppError{
		Error:   err,
		Message: fmt.Sprintf(format, v...),
		Code:    500,
	}
}

type status AppError

func writeJSON(w http.ResponseWriter, v interface{}) (int, error) {
	w.Header().Set("Content-Type", "application/json")
	payload, _ := json.Marshal(v)
	return w.Write([]byte(payload))
}

// AuthTokenHandler verifies the passed token with firebase
// returns a user object as a JSON if succcessful
func AuthTokenHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	val, ok := auth.VerifyAuthToken(r)
	payload, _ := json.Marshal(val)
	if !ok {
		writeJSON(w, struct {
			Valid bool
			Err   string
		}{
			ok,
			string(payload),
		})
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(payload)
	}

}

func getAuthenticatedUser(w http.ResponseWriter, r *http.Request) (github.User, error) {
	payload, ok := auth.VerifyAuthToken(r)
	user := payload.User
	if !ok {
		err := fmt.Errorf("Auth Error")
		return user, err
	}
	return user, nil
}
