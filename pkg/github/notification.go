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
	"fmt"
	"time"

	"google.golang.org/appengine/log"

	"golang.org/x/net/context"
)

// Notification stores logging information of outgoing notifications
type Notification struct {
	ID        uint      `gorm:"primary_key;AUTO_INCREMENT"`
	UserID    uint64    `gorm:"index;not null;"`
	Email     string    `gorm:"not null;"`
	Type      Frequency `gorm:"not null;"`
	Repos     string    `gorm:"not null;type:TEXT;"`
	CreatedAt time.Time
}

type repoJSON struct {
	Repo   string
	Issues uint64
}

// AddNotification saves notification information for the user
func (u User) AddNotification(
	ctx context.Context,
	email string, emailType Frequency, data []Payload) {

	if DB == nil {
		log.Errorf(ctx, "Failed to save notification, invalid DB Connection")
	}
	if DB.First(&User{}, "id = ?", u.ID).RecordNotFound() == false {
		notif := Notification{
			UserID: u.ID,
			Email:  email,
			Type:   emailType,
		}
		repos := []repoJSON{}
		// item holds one subscription's email payload
		for _, item := range data {
			r := repoJSON{
				Repo: item.RepoName,
			}
			UpdateRepo(ctx, item.RepoName)
			repoData, err := u.GetRepos(item.RepoName)
			if err != nil || len(repoData) != 1 {
				log.Errorf(ctx, "Failed to save notification, invalid data for repository: %v", err)
				continue
			}
			r.Issues = repoData[0].IssuesOpen
			repos = append(repos, r)
		}
		repoString, err := json.Marshal(repos)
		if err != nil {
			log.Errorf(ctx, "Failed to save notification, error converting to JSON: %v", err)
		} else {
			notif.Repos = string(repoString)
			if DB.Create(&notif).Error != nil {
				log.Errorf(ctx, "Failed to save notification, DB Error: %v", err)
			}
		}
	}
}

// GetNotifications returns all notifications sent to a user if emailType is 0
// or all notifications of a particular emailType (daily/weekly/monthly)
func (u User) GetNotifications(emailType Frequency) ([]Notification, error) {
	results := []Notification{}
	if DB == nil {
		return nil, fmt.Errorf("Failed to get notifications, invalid DB Connection")
	}
	if emailType == 0 {
		// Get all notifications
		err := DB.Find(&results, "user_id = ?", u.ID).Error
		return results, err
	}
	// Get notifications of emailType
	err := DB.Find(&results, "user_id = ? AND Type = ?", u.ID, emailType).Error
	return results, err
}
