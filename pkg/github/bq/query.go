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

// Package bq provides definitions and methods for structs used for querying
// data from the a dataset on BigQuery
package bq

import (
	"golang.org/x/net/context"

	"cloud.google.com/go/bigquery"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// queryFetcher fetches results from bigquery using the queryString
func queryFetcher(ctx context.Context, queryStr string) (*bigquery.RowIterator, error) {

	projectID := appengine.AppID(ctx)
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	query := client.Query(queryStr)
	query.QueryConfig.UseStandardSQL = false
	query.QueryConfig.DisableQueryCache = false
	results, err := query.Read(ctx)
	if err != nil {
		log.Errorf(ctx, "Query: %v", queryStr)
		log.Errorf(ctx, "BigQuery Error:"+err.Error())
		return nil, err
	}
	return results, err
}

// Fetch executes a BigQuery Job using the supplied context, SelectBuilder
// and returns a row iterator with results.
func Fetch(ctx context.Context, sb SelectBuilder) (*bigquery.RowIterator, error) {
	queryStr, err := sb.SQL()
	if err != nil {
		return nil, err
	}
	return queryFetcher(ctx, queryStr)
}
