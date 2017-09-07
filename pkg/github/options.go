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
	"fmt"
	"time"
)

// Options allows for additional configuration of the query
type Options struct {
	Tables       []string // To override the default tables - when extending the query for different dates
	Repositories []string // To set repositories to fetch Issues From
	Kind         []string // To set the Kinds of Event to select - eg "opened", "closed",etc
	Order        []string // To override default fields for use in OrderBy Clause
	Limit        uint64   // To override default limit value
	Conditions   []string // To override all conditions for the Query
}

// getTables returns set tables or the default table for today's data
func (o *Options) getTables() []string {
	// return set tables if present
	if len(o.Tables) != 0 {
		return o.Tables
	}
	// returns default table in other cases
	// By default uses today's githubarchive table
	usLoc, _ := time.LoadLocation("America/Los_Angeles")
	today := time.Now().In(usLoc).Format("20060102")
	return []string{"githubarchive.day." + today}
}

// getOrder returns the set of fields to use to sort the result by or the default sort fields
func (o *Options) getOrder() []string {
	// return set fields if present
	if len(o.Order) != 0 {
		return o.Order
	}
	//else sort by descending order of Ids followed by author by default
	return []string{"id DESC", "author"}
}

// getLimits returns the set limit for number of results or the default limit
func (o *Options) getLimits() uint64 {
	// Return set limit if valid
	if o.Limit > 0 {
		return o.Limit
	}
	// Return default limit otherwise
	return 10
}

// SetTables sets the default table values for the given frequency type
// If an invalid frequency is passed, it returns the default table for today's github events
func (o *Options) SetTables(f Frequency) {

	usLoc, _ := time.LoadLocation("America/Los_Angeles")
	today := time.Now().In(usLoc)
	if f == Weekly {
		weekStart := today.AddDate(0, 0, -7).Format("2006-01-02")
		weekEnd := today.Format("2006-01-02")
		week := weekTableRange(weekStart, weekEnd)
		o.Tables = []string{week}
	} else if f == Monthly {
		month := subTable("month") + today.Format("200601")
		o.Tables = []string{month}
	} else {
		day := subTable("day") + today.Format("20060102")
		o.Tables = []string{day}
	}
}

func weekTableRange(start, end string) string {
	return fmt.Sprintf(`(TABLE_DATE_RANGE([%s],TIMESTAMP('%s'),TIMESTAMP('%s')))`,
		subTable("day"), start, end)
}

func subTable(subTable string) string {
	return "githubarchive." + subTable + "."
}
