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
	"testing"
)

var queryTests = []struct {
	testcase  string
	query     SelectBuilder // SelectBuilder chain for the query
	want      string        // Expected output
	wantError string        // Expected Error

}{
	{
		"Case: Select field, From",
		Select(Columns{{"field", "label"}}).From("table_name"),
		"SELECT field label FROM table_name",
		"",
	},
	{
		"Case: Select JSON_EXTRACT_Scalar path, From",
		Select(Columns{{"key.subkey", "label"}}, "jsonkey").From("table_name"),
		"SELECT JSON_EXTRACT_SCALAR(jsonkey,'$.key.subkey') label FROM table_name",
		"",
	},
	{
		"Case: Select * From",
		SelectAll().From("table_name"),
		"SELECT * FROM table_name",
		"",
	},
	{
		"Case: Select * From table_name Limit 10",
		SelectAll().From("table_name").Limit(10),
		"SELECT * FROM table_name LIMIT 10",
		"",
	},
	{
		"Case: Select * From table_name Order By f",
		SelectAll().From("table_name").OrderBy("f"),
		"SELECT * FROM table_name ORDER BY f",
		"",
	},
	{
		"Case: Select * From table_name WHERE x IN ('1','2','3')",
		SelectAll().From("table_name").Where(In("x", "1", "2", "3")),
		"SELECT * FROM table_name WHERE x IN ('1','2','3')",
		"",
	},
	{
		"Case: Select with JSON Extracted field",
		SelectAll().
			From("user_favorites").
			Where(JExtract("favorite", "movie.title") + " IN ('Hello World')"),
		`SELECT * FROM user_favorites WHERE JSON_EXTRACT_SCALAR(favorite,'$.movie.title') ` +
			`IN ('Hello World')`,
		"",
	},
	{
		"Case: Select f1,f2,f3 From table_name Where c1 AND c2 OR c3",
		Select(Columns{{"f1", "l1"}}).Select(Columns{{"f2", "l2"}, {"f3", ""}}).
			From("table_name").
			Where("f1 IN ('xyz','qwerty')").
			And("f2 LIKE ('%asdf%')").Or("f3 = '@'"),
		`SELECT f1 l1, f2 l2, f3 FROM table_name WHERE f1 IN ('xyz','qwerty') AND f2 LIKE ` +
			`('%asdf%') OR f3 = '@'`,
		"",
	},
	{
		"ErrorCase: for Empty From Clause",
		SelectAll().From(""),
		"",
		"Select statement must have at least one target table",
	},
	{
		"ErrorCase: for Missing From Clause",
		SelectAll().Where("XYZ > 10"),
		"",
		"Select statement must have at least one target table",
	},
}

//TestSelectBuilder runs a series of tests on SelectBuilder which are expected to pass
func TestSelectBuilder(t *testing.T) {
	for _, tt := range queryTests {
		sql, err := tt.query.SQL()
		if err != nil && err.Error() != tt.wantError {
			t.Errorf("\n%v\nSelectBuilder.SQL() failed with error: \n %v\ninsead of error:\n %v",
				tt.testcase, err, tt.wantError)
		} else if sql != tt.want {
			t.Errorf("\n%v\nSelectBuilder.SQL() gave a wrong result:\nWant:\n %v\nGot:\n %v\n",
				tt.testcase, tt.want, sql)
		}
	}
}
