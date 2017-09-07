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

// Package db provides types and methods database operations.
package db

import (
	"log"
	"os"

	"google.golang.org/appengine"

	"github.com/jinzhu/gorm"
)

type cloudSQLConfig struct {
	Username, Password string
}

func configureCloudSQL(config cloudSQLConfig) (*gorm.DB, error) {

	if appengine.IsDevAppServer() {
		// Running locally.
		return newORMClient(MySQLConfig{
			Username: config.Username,
			Password: config.Password,
			Host:     "localhost",
			Port:     3306,
		})
	}

	// Running in production.
	return newORMClient(MySQLConfig{
		Username: config.Username,
		Password: config.Password,
		Instance: mustGetenv("CLOUDSQL_CONNECTION_NAME"),
	})

}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Panicf("%s environment variable not set.", k)
	}
	return v
}

// GetConnection returns the configured db connection or nil
func GetConnection() *gorm.DB {
	var err error
	if connection == nil {
		connection, err = configureCloudSQL(cloudSQLConfig{
			Username: mustGetenv("CLOUDSQL_USER"),
			Password: mustGetenv("CLOUDSQL_PASSWORD"),
		})
		if err != nil || connection == nil {
			log.Fatalf("Error connecting: %v", err)
		}
		return connection
	}
	return connection
}
