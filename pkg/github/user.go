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
	"io/ioutil"
	"time"

	"golang.org/x/net/context"
)

const (
	repoAPI = "repos/"
)

// Frequency is used for tracking notification frequency preferences
type Frequency int

// Named frequency constants to help make code readable
const (
	_                 = iota // skip 0 value
	Never   Frequency = iota // 1
	Daily                    // 2
	Weekly                   // 3
	Monthly                  // 4
)

// EmailPreference stores frequency for various types of notifications
type EmailPreference struct {
	SubscriptionID uint      `gorm:"unique;index;not null;"`
	IssueOpen      Frequency `gorm:"type:INT;" sql:"DEFAULT:1"`
	IssueClose     Frequency `gorm:"type:INT;" sql:"DEFAULT:1"`
	IssueReopen    Frequency `gorm:"type:INT;" sql:"DEFAULT:1"`
	NewComment     Frequency `gorm:"type:INT;" sql:"DEFAULT:1"`
	NoComment      Frequency `gorm:"type:INT;" sql:"DEFAULT:1"`
}

// Repo stores open issue count for repositories
type Repo struct {
	ID         uint64 `gorm:"primary_key;AUTO_INCREMENT"`
	Name       string `gorm:"index;not null;"`
	IssuesOpen uint64 `gorm:"type:INT;" `
	UpdatedAt  time.Time
}

// NewPreference returns a default email frequency for all types set to daily
func NewPreference() EmailPreference {
	return EmailPreference{
		IssueOpen:   Daily,
		IssueClose:  Daily,
		IssueReopen: Daily,
		NewComment:  Daily,
		NoComment:   Daily,
	}
}

// Subscription stores information that relates a repo with a user's watch list
type Subscription struct {
	ID                 uint            `gorm:"primary_key;AUTO_INCREMENT"`
	UserID             uint64          `gorm:"index;not null;"`
	Repo               string          `gorm:"index;not null;"`
	DefaultEmail       string          `gorm:"not null;"`
	EmailPreference    EmailPreference `gorm:"ForeignKey:SubscriptionID"`
	LastNotificationID uint64          // Last notification’s ID for sending reminders if needed

}

// User stores basic user data
type User struct {
	ID            uint64         `gorm:"primary_key;"`  // Unique Identifier for user Entity
	FireKey       string         `gorm:"unique_index;"` // User's firebase UID
	Login         string         `gorm:"unique_index;"` // Github ID for the user
	Email         string         `gorm:"unique_index;"` // User's default email
	Subscriptions []Subscription `gorm:"ForeignKey:UserID"`
	CreatedAt     time.Time
}

//Add inserts a user record for the calling object to the DB
func (u *User) Add() error {
	if DB == nil {
		return fmt.Errorf("Failed to Add user, invalid DB Connection")
	}
	if DB.First(&User{}, "id = ?", u.ID).RecordNotFound() {
		if err := DB.Create(&u).Error; err != nil {
			return err
		}
		// update the user object
		DB.First(&u, "id = ?", u.ID)
		return nil
	}

	return fmt.Errorf("Failed to Add user, duplicate record")
}

// IsNew returns true if there is an entry for the given user in the database
func (u *User) IsNew() (User, bool) {
	var user User
	if DB.Preload("Subscriptions").Preload("Subscriptions.EmailPreference").
		First(&user, "id = ?", u.ID).RecordNotFound() {
		return User{}, true
	}
	return user, false
}

//Update updates the user record for the calling object or creates a new record in the DB
func (u *User) Update() error {
	if DB == nil {
		return fmt.Errorf("Failed to update user, invalid DB Connection")
	}
	if DB.First(&User{}, "id = ?", u.ID).RecordNotFound() == false {
		updates := *u
		updates.Subscriptions = []Subscription{}
		return DB.Save(&updates).Error
	}

	return fmt.Errorf("Failed to Update user, record not found")
}

//Remove deletes the user record for the calling object to the DB if it exists
func (u User) Remove() error {
	if DB == nil {
		return fmt.Errorf("Failed to remove user, invalid DB Connection")
	}

	if u.ID == 0 || DB.First(&User{}, "id = ?", u.ID).RecordNotFound() {
		return fmt.Errorf("Failed to remove user, invalid ID")
	}

	if err := DB.Delete(&u).Error; err != nil {
		return err
	}
	return nil
}

// Subscribe adds a subscription to a user
//
// Optional argument is DefaultEmail
func (u *User) Subscribe(repo string, pref EmailPreference, argv ...string) error {
	sub := Subscription{
		UserID:          u.ID,
		Repo:            repo,
		EmailPreference: pref,
		DefaultEmail:    u.Email,
	}
	if argc := len(argv); argc == 0 {
		sub.DefaultEmail = u.Email
	} else if argc == 1 && len(argv[0]) != 0 {
		sub.DefaultEmail = argv[0]
	}

	if DB == nil {
		return fmt.Errorf("Failed to subscribe, invalid DB Connection")
	}
	if DB.First(&u, "id = ?", u.ID).RecordNotFound() == false {
		if existing, _ := u.GetSubscriptions(repo); existing == nil {
			u.Subscriptions = append(u.Subscriptions, sub)
			if err := DB.Create(&sub).Error; err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("Failed to subscribe, subscription already exists")
	}
	return fmt.Errorf("Failed to subscribe, no such user")
}

// GetSubscriptions returns all the subscriptions that match passed repos for a user
// If no repos are passed, it returns the entire list of subscriptions
func (u User) GetSubscriptions(repos ...string) ([]Subscription, error) {
	results := []Subscription{}
	if DB == nil {
		return nil, fmt.Errorf("Failed to get subscriptions, invalid DB Connection")
	}
	if len(repos) == 0 {
		// Get all subscriptions
		err := DB.Preload("EmailPreference").Find(&results, "user_id = ?", u.ID).Error
		return results, err
	}
	for _, repo := range repos {
		var sub Subscription
		if DB.Preload("EmailPreference").Where("user_id = ? AND repo = ?", u.ID, repo).First(&sub).RecordNotFound() == false {
			results = append(results, sub)
		} else {
			return nil, fmt.Errorf("No such subscription: %s", repo)
		}
	}
	return results, nil
}

// UnsubscribeAll removes all subscriptions for a user
func (u *User) UnsubscribeAll() error {
	subs, err := u.GetSubscriptions()
	if err != nil {
		return err
	}
	err = DB.Where("user_id = ?", u.ID).Delete(Subscription{}).Error
	u.Subscriptions = unsubscriber(u.Subscriptions, subs)
	return err
}

// Unsubscribe removes a subscription from a user
func (u *User) Unsubscribe(repos ...string) error {
	if DB == nil {
		return fmt.Errorf("Failed to unsubscribe, invalid DB Connection")
	}
	var err error
	if err = DB.First(u).Error; err == nil {
		var subs []Subscription
		for _, repo := range repos {
			var sub Subscription
			if DB.Where("user_id = ? AND repo = ?", u.ID, repo).First(&sub).RecordNotFound() == false {
				if DB.Delete(&sub).Error != nil {
					return fmt.Errorf("Failed to unsubscribe:%v", sub.ID)
				}
				subs = append(subs, sub)
			}
		}
		u.Subscriptions = unsubscriber(u.Subscriptions, subs)
		return nil

	}
	return fmt.Errorf("Failed to unsubscribe, %v", err)
}

// UpdateSubscription updates a user's subscription preferences
func (u *User) UpdateSubscription(repo string, s *Subscription) error {
	subs, err := u.GetSubscriptions(repo)
	if err != nil {
		return err
	}
	sub := subs[0]
	DB.Delete(&EmailPreference{}, "subscription_id = ?", sub.ID)
	s.EmailPreference.SubscriptionID = sub.ID
	err = DB.Save(&s.EmailPreference).Error
	if err != nil {

		return fmt.Errorf("Failed to update preferences: %v", err)
	}
	return DB.Model(&sub).UpdateColumns(s).Error
}

// UpdateRepo creates or updates repo data for a repository given by r from Github
func UpdateRepo(ctx context.Context, r string) error {
	var repo Repo
	if DB == nil {
		return fmt.Errorf("Failed to update repo, invalid DB Connection")
	}
	// Make a request to Github's API for repo data
	resp, err := API(ctx, repoAPI+r)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("API Error: %s", resp.Status)
	}
	resBody, _ := ioutil.ReadAll(resp.Body)
	var jsonResponse map[string]interface{}
	json.Unmarshal(resBody, &jsonResponse)
	count, _ := jsonResponse["open_issues"].(float64)
	openIssues := uint64(count)
	if DB.First(&repo, "name = ?", r).RecordNotFound() == false {
		repo.IssuesOpen = openIssues
		return DB.Save(&repo).Error
	}
	repo.Name = r
	repo.IssuesOpen = openIssues
	return DB.Create(&repo).Error
}

// GetRepos returns data for  passed repos
// If no repos are passed, it returns the entire list of repo data stored for a user
func (u User) GetRepos(repos ...string) ([]Repo, error) {
	results := []Repo{}
	if DB == nil {
		return nil, fmt.Errorf("Failed to get repositories, invalid DB Connection")
	}
	if len(repos) == 0 {
		// Get all subscriptions
		userRepos := []string{}
		for _, sub := range u.Subscriptions {
			userRepos = append(userRepos, sub.Repo)
		}
		err := DB.Model(&Repo{}).Where("name in (?)", userRepos).Find(&results).Error
		return results, err
	}
	for _, name := range repos {
		var repo Repo
		if DB.Where("name = ?", name).First(&repo).RecordNotFound() == false {
			results = append(results, repo)
		}
	}
	return results, nil
}

//FindUserByLogin retrieves a user object by their Github ID
func FindUserByLogin(login string) (User, error) {
	var user User
	if DB == nil {
		return user, fmt.Errorf("Failed to FindUserByID, invalid DB Connection")
	}
	if err := DB.Preload("Subscriptions").Preload("Subscriptions.EmailPreference").First(&user, "login = ?", login).Error; err != nil {
		return user, err
	}
	return user, nil
}

//GetUsers retrieves all user objects
func GetUsers() ([]User, error) {
	var users []User
	if DB == nil {
		return nil, fmt.Errorf("Failed to GetUsers, invalid DB Connection")
	}
	if err := DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// unsubscriber is a helper method to remove subscriptions
func unsubscriber(source []Subscription, toRemove []Subscription) []Subscription {
	mb := map[uint]bool{}
	for _, x := range toRemove {
		mb[x.ID] = true
	}
	result := []Subscription{}
	for _, x := range source {
		if _, ok := mb[x.ID]; !ok {
			result = append(result, x)
		}
	}
	return result
}
