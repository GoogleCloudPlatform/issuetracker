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

package bq

import (
	"bytes"
	"fmt"

	"strings"

	"github.com/lann/builder"
)

// Pair stores key value pairs that are used in various methods of the builder
//
// The syntax results in a statement like "Select html_url as url" in traditional SQL
type Pair struct {
	Key   string
	Value string
}

// Columns Stores an array of Column Name,Label Pairs.
type Columns []Pair

// selectComponents builds a BigQuery query string for executing a select query on BigQuery
type selectComponents struct {
	Fields     []string // holds the set of fields for use with select clause
	Tables     []string // holds the list of tables to query from
	Conditions []string // holds the list of conditions to use
	OrderBys   []string // field to apply Order By clause
	Limit      string   // Limit number of rows in result set
}

// SQL returns a formatted query for use with BigQuery's Legacy SQL Dialect
func (s *selectComponents) SQL() (query string, err error) {
	if len(s.Fields) == 0 {
		err := fmt.Errorf("Select statement must have at least one field")
		return "", err
	}
	if len(s.Tables) == 0 {
		err := fmt.Errorf("Select statement must have at least one target table")
		return "", err
	}
	rawQuery := &bytes.Buffer{}

	// Add column(s)
	rawQuery.WriteString("SELECT ")
	rawQuery.WriteString(strings.Join(s.Fields, ", "))

	// Add table(s)
	rawQuery.WriteString(" FROM ")
	rawQuery.WriteString(strings.Join(s.Tables, ", "))

	// Add where clauses
	if len(s.Conditions) > 0 {
		rawQuery.WriteString(" WHERE ")
		rawQuery.WriteString(strings.Join(s.Conditions, " "))
	}

	// Add Order By clauses
	if len(s.OrderBys) > 0 {
		rawQuery.WriteString(" ORDER BY ")
		rawQuery.WriteString(strings.Join(s.OrderBys, ", "))
	}

	// Add Limit clause
	if len(s.Limit) > 0 {
		rawQuery.WriteString(" LIMIT ")
		rawQuery.WriteString(s.Limit)
	}

	return rawQuery.String(), nil
}

// SelectBuilder is an exported Builder that allows users to build Legacy SQL queries on BigQuery
type SelectBuilder builder.Builder

func init() {
	builder.Register(SelectBuilder{}, selectComponents{})
}

// Select adds field and label from
// Columns{Pair{Key: field, Value: label)} to the list columns to select
//
// For Example: Select({'html_url': 'url'}) results in  SELECT html_url  url
//
// Select also supports JSON Extracted field names - by passing the path as
// Column key and JSON field name as second argument
func (b SelectBuilder) Select(cols Columns, args ...string) SelectBuilder {

	var fields []string
	for _, col := range cols {
		field := col.Key
		if len(args) == 1 {
			//use JSON_EXTRACT_SCALAR Helper
			field = JExtract(args[0], col.Key)
		}
		if len(col.Value) != 0 {
			field = field + " " + col.Value
		}
		fields = append(fields, field)
	}
	return builder.Extend(b, "Fields", fields).(SelectBuilder)
}

// SelectAll sets Fields in the selectComponents to "*", which results in a
// "SELECT * FROM..." like query
func (b SelectBuilder) SelectAll() SelectBuilder {
	return builder.Set(b, "Fields", []string{"*"}).(SelectBuilder)
}

// From adds all strings in names to use as tables in the query
func (b SelectBuilder) From(names ...string) SelectBuilder {
	var tables []string
	for _, s := range names {
		if len(s) != 0 {
			tables = append(tables, s)
		}
	}
	return builder.Extend(b, "Tables", tables).(SelectBuilder)
}

// Where adds condition to the b.Condition.
// If previous conditions exist, it behaves like b.And(condition)
func (b SelectBuilder) Where(condition string) SelectBuilder {
	if _, ok := builder.Get(b, "Conditions"); ok {
		// Other conditions exists, so we append with And Clause
		return b.And(condition)
	}
	return builder.Extend(b, "Conditions", []string{condition}).(SelectBuilder)
}

// addCondition adds all strings in clauses to b.Conditions with the given condition operator
func (b SelectBuilder) addCondition(operator string, clauses ...string) SelectBuilder {

	if _, ok := builder.Get(b, "Conditions"); !ok {
		// When no condition has been added, omit the operator for first condition
		return b.Where(clauses[0]).addCondition(operator, clauses[1:]...)
	}

	for i, s := range clauses {
		if len(s) != 0 {
			clauses[i] = operator + s
		}
	}
	return builder.Extend(b, "Conditions", clauses).(SelectBuilder)
}

// And adds all strings in conditions to b.Conditions with an "AND" prefix
func (b SelectBuilder) And(conditions ...string) SelectBuilder {
	return b.addCondition("AND ", conditions...)
}

// listToStr is a helper method used by In and Like methods
// converts strings x,y,z.. in list to ('x','y','z') and adds it to prefix
func listToStr(prefix string, list ...string) string {
	for i, s := range list {
		list[i] = "'" + s + "'"
	}
	return prefix + " (" + strings.Join(list, ",") + ")"
}

// In can be used to write conditions of form "field IN (list of values)"
func In(field string, values ...string) string {
	return listToStr(field+" IN", values...)
}

// NotIn can be used to write conditions of form "field NOT IN (list of values)"
func NotIn(field string, values ...string) string {
	return listToStr(field+" NOT IN", values...)
}

// Like can be used to write conditions form "field LIKE (list of values)"
func Like(field string, values ...string) string {
	return listToStr(field+" LIKE", values...)
}

// NotLike can be used to write conditions of form "field NOT LIKE (list of values)"
func NotLike(field string, values ...string) string {
	return listToStr(field+" NOT LIKE", values...)
}

// Or adds all strings in conditions to b.Conditions with an "OR" prefix
func (b SelectBuilder) Or(conditions ...string) SelectBuilder {
	return b.addCondition("OR ", conditions...)
}

// OrderBy adds all strings in fields to b.SortFields separated by spaces
func (b SelectBuilder) OrderBy(fields ...string) SelectBuilder {
	return builder.Extend(b, "OrderBys", fields).(SelectBuilder)
}

// Limit adds a limit clause to the query
func (b SelectBuilder) Limit(limit uint64) SelectBuilder {
	//fmt.Print(debug(b))
	return builder.Set(b, "Limit", fmt.Sprintf("%d", limit)).(SelectBuilder)
}

// SQL composes the query string to use with BigQuery's Legacy SQL standard
func (b SelectBuilder) SQL() (string, error) {
	components := builder.GetStruct(b).(selectComponents)
	return components.SQL()
}

// IsEmpty checks if the builder has any fields and returns true or false
func (b SelectBuilder) IsEmpty() bool {
	components := builder.GetStruct(b).(selectComponents)
	return len(components.Fields) == 0 && len(components.Tables) == 0
}

// JExtract formats a Column in a select query to use JSON_EXTRACT_SCALAR
// As per https://cloud.google.com/bigquery/docs/reference/legacy-sql#json_extract_scalar
//
// Example: JExtract('payload','action') ==> JSON_EXTRACT_SCALAR(payload,'$.action')
func JExtract(json string, path string) string {
	return fmt.Sprintf("JSON_EXTRACT_SCALAR(%s,'$.%s')", json, path)
}

// Select acts as wrapper to easily start a new SelectBuilder Chain
func Select(c Columns, args ...string) SelectBuilder {
	return SelectBuilder{}.Select(c, args...)
}

// SelectAll functions similar to the Select Wrapper
func SelectAll() SelectBuilder {
	return SelectBuilder{}.SelectAll()
}

func Debug(t SelectBuilder) SelectBuilder {
	fmt.Print(t.SQL())
	return t
}
