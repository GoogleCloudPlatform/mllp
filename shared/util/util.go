// Copyright 2018 Google LLC
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

// Package util provides parsing and formatting for the paths used by the REST API.
package util

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	goauth2 "golang.org/x/oauth2/google"
	"golang.org/x/oauth2"
)

const (
	// Common URL path components.
	projectsPathComponent    = "projects"
	locationsPathComponent   = "locations"
	datasetsPathComponent    = "datasets"
	hl7V2StoresPathComponent = "hl7V2Stores"
	messagesPathComponent    = "messages"
	fhirStoresPathComponent  = "fhirStores"
)

// GenerateHL7V2StoreName puts together the components to form the name of a REST HL7v2 store resource.
func GenerateHL7V2StoreName(projectReference, locationID, datasetID, hl7V2StoreID string) string {
	return strings.Join([]string{
		projectsPathComponent,
		projectReference,
		locationsPathComponent,
		locationID,
		datasetsPathComponent,
		datasetID,
		hl7V2StoresPathComponent,
		hl7V2StoreID,
	}, "/")
}

// GenerateHL7V2MessageName puts together the components to form the name of a REST Message resource.
func GenerateHL7V2MessageName(projectReference, locationID, datasetID, hl7V2StoreID, messageID string) string {
	return strings.Join([]string{
		GenerateHL7V2StoreName(projectReference, locationID, datasetID, hl7V2StoreID),
		messagesPathComponent,
		messageID,
	}, "/")
}

// GenerateFHIRStoreName puts together the components to form the name of a REST FHIR store.
func GenerateFHIRStoreName(projectReference, locationID, datasetID, fhirStoreID string) string {
	return strings.Join([]string{
		projectsPathComponent,
		projectReference,
		locationsPathComponent,
		locationID,
		datasetsPathComponent,
		datasetID,
		fhirStoresPathComponent,
		fhirStoreID,
	}, "/")
}

// ParseHL7V2MessageName parses the project reference, location id, dataset id,
// HL7v2 store id, and message id from the given resource name.
func ParseHL7V2MessageName(name string) (string, string, string, string, string, error) {
	parts := strings.Split(name, "/")
	ids, i := []string{}, 0
	allComponents := []string{projectsPathComponent, locationsPathComponent, datasetsPathComponent, hl7V2StoresPathComponent, messagesPathComponent}
	for _, component := range allComponents {
		if len(component) != 0 {
			if len(parts) <= i || parts[i] != component {
				return "", "", "", "", "", fmt.Errorf("expected component %v at position %v in %v", component, i, parts)
			}
			i++
		}
		if len(parts) <= i {
			return "", "", "", "", "", fmt.Errorf("expected a component at position %v in %v", i, parts)
		}
		ids = append(ids, parts[i])
		i++
	}
	if len(parts[i:]) > 0 {
		return "", "", "", "", "", fmt.Errorf("unexpected tokens %v in %v", parts[i:], name)
	}

	return ids[0], ids[1], ids[2], ids[3], ids[4], nil
}

// TokenSource creates a token source for authenticating against GCP services.
// If a credentials file is provided (e.g. for a customized service account),
// it will be used to generate the token source, otherwise the default service
// account will be used.
func TokenSource(ctx context.Context, cred string, scopes ...string) (oauth2.TokenSource, error) {
	if cred == "" {
		return goauth2.DefaultTokenSource(ctx, scopes...)
	}
	j, err := ioutil.ReadFile(cred)
	if err != nil {
		return nil, err
	}
	c, err := goauth2.CredentialsFromJSON(ctx, j, scopes...)
	if err != nil {
		return nil, err
	}
	fmt.Printf("***** %v", c.TokenSource)
	return c.TokenSource, nil
}
