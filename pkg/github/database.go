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
	"log"

	"github.com/GoogleCloudPlatform/issuetracker/pkg/github/db"

	"github.com/jinzhu/gorm"
)

// DB exposes the connection to the backend database
var (
	DB *gorm.DB
)

func init() {
	DB = db.GetConnection()
	if DB == nil {
		log.Panicf("Error migrating,  Invalid database connection")
	}
	err := DB.AutoMigrate(
		&User{},
		&Repo{},
		&Subscription{},
		&EmailPreference{},
		&Notification{},
	).Error

	if err != nil {
		log.Panicf("Error migrating tables,%v", err)
	} else {
		DB.Model(&Subscription{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
		DB.Model(&Subscription{}).AddUniqueIndex("idx_userid_repo", "user_id", "repo")
		DB.Model(&EmailPreference{}).
			AddForeignKey("subscription_id", "subscriptions(id)", "CASCADE", "CASCADE")
		DB.Model(&Notification{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")

	}
}
