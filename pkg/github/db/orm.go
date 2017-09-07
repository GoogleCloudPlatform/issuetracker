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

package db

import (
	"fmt"
	// Import the mysql drivers for Gorm
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// connection holds the orm handle
var connection *gorm.DB

// MySQLConfig holds SQL configuration information for setting up connections
type MySQLConfig struct {
	// Optional.
	Username, Password string

	// Host of the MySQL instance.
	//
	// If set, Instance should be unset.
	Host string

	// Port of the MySQL instance.
	//
	// If set, Instance should be unset.
	Port int

	// Instance is the CloudSQL instance name.
	//
	// If set, Host and Port should be unset.
	Instance string
}

// dataStoreName returns a connection string suitable for sql.Open.
func (c MySQLConfig) dataStoreName(databaseName string) string {
	var cred string
	// [username[:password]@]
	if c.Username != "" {
		cred = c.Username
		if c.Password != "" {
			cred = cred + ":" + c.Password
		}
		cred = cred + "@"
	}

	if c.Instance != "" {
		return fmt.Sprintf("%s:%s@cloudsql(%s)/%s?parseTime=true", c.Username, c.Password,
			c.Instance, databaseName)
	}
	return fmt.Sprintf("%stcp([%s]:%d)/%s?parseTime=true", cred, c.Host, c.Port, databaseName)
}

// newORMClient creates a connection to a CloudSQL/MySQL instance using the provided credentials
// and stores the connection object in db. "ghdata" is the name of the database
func newORMClient(config MySQLConfig) (*gorm.DB, error) {

	db, err := gorm.Open("mysql", config.dataStoreName("ghdata"))
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	return db, nil
}
