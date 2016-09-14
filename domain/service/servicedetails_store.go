// Copyright 2016 The Serviced Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"errors"
	"strings"

	"github.com/control-center/serviced/datastore"
	"github.com/control-center/serviced/datastore/elastic"
)

// GetAllServiceDetails returns service details for an id
func (s *storeImpl) GetAllServiceDetails(ctx datastore.Context) ([]ServiceDetails, error) {
	searchRequest := newServiceDetailsElasticRequest(map[string]interface{}{
		"query": map[string]interface{}{
			"query_string": map[string]string{
				"query": "_exists_:ID",
			},
		},
		"fields": serviceDetailsFields,
		"size":   50000,
	})

	results, err := datastore.NewQuery(ctx).Execute(searchRequest)
	if err != nil {
		return nil, err
	}

	details := []ServiceDetails{}
	for results.HasNext() {
		var d ServiceDetails
		err := results.Next(&d)
		if err != nil {
			return nil, err
		}
		details = append(details, d)
	}

	return details, nil
}

// GetServiceDetails returns service details for an id
func (s *storeImpl) GetServiceDetails(ctx datastore.Context, serviceID string) (*ServiceDetails, error) {
	id := strings.TrimSpace(serviceID)
	if id == "" {
		return nil, errors.New("empty service id not allowed")
	}

	searchRequest := newServiceDetailsElasticRequest(map[string]interface{}{
		"query": map[string]interface{}{
			"ids": map[string]interface{}{
				"values": []string{id},
			},
		},
		"fields": serviceDetailsFields,
		"size":   50000,
	})

	results, err := datastore.NewQuery(ctx).Execute(searchRequest)
	if err != nil {
		return nil, err
	}

	if results.HasNext() {
		var details ServiceDetails
		err = results.Next(&details)
		if err != nil {
			return nil, err
		}
		return &details, nil
	}

	// no errors but also no results for given id
	return nil, nil
}

// GetChildServiceDetails returns service details given parent service id
func (s *storeImpl) GetServiceDetailsByParentID(ctx datastore.Context, parentID string) ([]ServiceDetails, error) {
	searchRequest := newServiceDetailsElasticRequest(map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]string{"ParentServiceID": parentID},
		},
		"fields": serviceDetailsFields,
		"size":   50000,
	})

	results, err := datastore.NewQuery(ctx).Execute(searchRequest)
	if err != nil {
		return nil, err
	}

	details := []ServiceDetails{}
	for results.HasNext() {
		var d ServiceDetails
		err := results.Next(&d)
		if err != nil {
			return nil, err
		}
		details = append(details, d)
	}

	return details, nil
}

func newServiceDetailsElasticRequest(query interface{}) elastic.ElasticSearchRequest {
	return elastic.ElasticSearchRequest{
		Pretty: false,
		Index:  "controlplane",
		Type:   "service",
		Scroll: "",
		Scan:   0,
		Query:  query,
	}
}

var serviceDetailsFields = []string{
	"ID",
	"Name",
	"Description",
	"PoolID",
	"ParentServiceID",
	"Instances",
	"InstanceLimits",
	"RAMCommitment",
	"Startup",
}