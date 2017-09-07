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
	"os"
	"strconv"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestORMClientConnection(t *testing.T) {
	t.Parallel()

	host := os.Getenv("MYSQL_HOST")
	user := os.Getenv("CLOUDSQL_USER")
	pass := os.Getenv("CLOUDSQL_PASSWORD")
	port := os.Getenv("MYSQL_PORT")

	if host == "" {
		t.Skip("MYSQL_HOST not set.")
	}
	if port == "" {
		port = "3306"
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("Could not parse port: %v", err)
	}

	db, err := newORMClient(MySQLConfig{
		Username: user,
		Password: pass,
		Host:     host,
		Port:     p,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

}
