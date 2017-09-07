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

// Package testutil provides helpers for configuring tests for the project
package testutil

import (
	"errors"
	"os"

	"google.golang.org/appengine/aetest"
)

var errProjectID = errors.New("PROJECT_ID not set")

// GetOptions returns options for aetest instances by setting the right project ID
func GetOptions() (aetest.Options, error) {
	ProjectID := os.Getenv("PROJECT_ID")
	if ProjectID == "" {
		return aetest.Options{}, errProjectID
	}
	return aetest.Options{
		AppID: ProjectID,
	}, nil
}
