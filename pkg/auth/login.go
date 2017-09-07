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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/github"

	firebase "github.com/wuman/firebase-server-sdk-go"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

const (
	userAPI = "user/"
)

// Status consists of the fields that will be returned on successful authentication
type Status struct {
	User    github.User
	Valid   bool
	Onboard bool
}

func init() {
	firebase.InitializeApp(&firebase.Options{
		ServiceAccountPath: CredentialsPath(),
	})
}

// VerifyAuthToken checks if the request contains a valid Firebase Auth token
func VerifyAuthToken(r *http.Request) (Status, bool) {

	//initialise the default return status
	s := Status{
		github.User{},
		false,
		false,
	}
	ctx := appengine.NewContext(r)
	auth, err := firebase.GetAuth()
	if err != nil {
		log.Errorf(ctx, "Credentials Error: %v", err)
		return s, false
	}
	token := r.Header.Get("Authorization")

	// We have to use urlfetch when using App Engine
	decodedToken, err := auth.VerifyIDTokenWithTransport(token, urlfetch.Client(ctx).Transport)
	if err != nil {
		log.Infof(ctx, "Credentials Error: %v", err)
		return s, false
	}
	claims := decodedToken.Claims()
	firebaseClaims, ok := claims["firebase"].(map[string]interface{})
	if !ok {
		log.Errorf(ctx, "Firebase Claims Error: %v", err)
		return s, false
	}
	identities, ok := firebaseClaims["identities"].(map[string]interface{})
	if !ok {
		log.Errorf(ctx, "Firebase Identities Error: %v", ok)
		return s, false
	}
	var u github.User
	idString := ""
	ID, ok := identities["github.com"].([]interface{})
	if !ok {
		log.Errorf(ctx, "Failed to get provider data: %v", ok)
		return s, false
	}
	idString = ID[0].(string)
	u.ID, _ = strconv.ParseUint(idString, 10, 64)

	email, ok := identities["email"].([]interface{})
	if !ok {
		log.Errorf(ctx, "Invalid email error: %v", ok)
		return s, false
	}
	u.Email, _ = email[0].(string)
	u.FireKey, ok = claims["user_id"].(string)
	if !ok {
		log.Errorf(ctx, "Firebase Key Error: %v", err)
		return s, false
	}
	userFromDB, newUser := u.IsNew()
	if newUser || userFromDB.ID == 0 {
		// Get login name from Github
		resp, err := github.API(ctx, userAPI+idString)
		if err != nil {
			return s, false
		}
		resBody, _ := ioutil.ReadAll(resp.Body)
		var jsonResponse map[string]interface{}
		json.Unmarshal(resBody, &jsonResponse)
		u.Login = jsonResponse["login"].(string)
		u.Add()
		s = Status{u, true, true}
	} else {
		s = Status{userFromDB, true, false}
	}

	// returns user, validity and whether or not the user is a new user
	return s, true
}

// CredentialsPath returns a canonical path to the credentials.json file
func CredentialsPath() string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	basepath = filepath.Join(basepath, "..", "..")
	return basepath + "/pkg/auth/credentials.json"
}
