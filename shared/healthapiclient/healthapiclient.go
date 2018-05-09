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

// Package healthapiclient handles communication with the APIs.
package healthapiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/golang/glog"
	"golang.org/x/net/context"
	"shared/monitoring"
	"shared/util"

	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	goauth2 "golang.org/x/oauth2/google"
)

const (
	scope               = "https://www.googleapis.com/auth/cloud-healthcare"
	contentType         = "application/json"
	fhirJSONContentType = "application/fhir+json;charset=utf-8"
	sendSuffix          = "messages:ingest"

	sentMetric               = "apiclient-sent"
	sendErrorMetric          = "apiclient-send-error"
	fetchedMetric            = "apiclient-fetched"
	fetchErrorMetric         = "apiclient-fetch-error"
	fetchErrorInternalMetric = "apiclient-fetch-error-internal"
)

// HL7Client represents a client of the HL7 API.
type HL7Client struct {
	metrics          *monitoring.Client
	client           *http.Client
	apiAddrPrefix    string
	projectReference string
	locationID       string
	datasetID        string
	hl7StoreID       string
}

// FHIRClient represents a client of the FHIR API.
type FHIRClient struct {
	metrics          *monitoring.Client
	client           *http.Client
	apiAddrPrefix    string
	projectReference string
	locationID       string
	datasetID        string
	fhirStoreID      string
}

type message struct {
	Data []byte `json:"data"`
}

type sendMessageReq struct {
	Msg message `json:"message"`
}

type sendMessageResp struct {
	Hl7Ack []byte `json:"hl7_ack"`
}

// NewHL7Client creates a properly authenticated client that talks to an HL7 backend.
func NewHL7Client(ctx context.Context, metrics *monitoring.Client, apiAddrPrefix, projectID, locationID, datasetID, hl7StoreID string) (*HL7Client, error) {
	if err := validatesComponents(projectID, locationID, datasetID, hl7StoreID); err != nil {
		return nil, err
	}

	httpClient, err := initHTTPClient(ctx, apiAddrPrefix)
	if err != nil {
		return nil, err
	}

	c := &HL7Client{
		metrics:          metrics,
		client:           httpClient,
		apiAddrPrefix:    apiAddrPrefix,
		projectReference: projectID,
		locationID:       locationID,
		datasetID:        datasetID,
		hl7StoreID:       hl7StoreID,
	}
	c.initMetrics()
	return c, nil
}

func (c *HL7Client) initMetrics() {
	c.metrics.NewInt64(sentMetric)
	c.metrics.NewInt64(sendErrorMetric)
	c.metrics.NewInt64(fetchedMetric)
	c.metrics.NewInt64(fetchErrorMetric)
	c.metrics.NewInt64(fetchErrorInternalMetric)
}

// NewFHIRClient creates a properly authenticated client that talks to a FHIR backend.
func NewFHIRClient(ctx context.Context, metrics *monitoring.Client, apiAddrPrefix, projectID, locationID, datasetID, fhirStoreID string) (*FHIRClient, error) {
	if err := validatesComponents(projectID, locationID, datasetID, fhirStoreID); err != nil {
		return nil, err
	}

	httpClient, err := initHTTPClient(ctx, apiAddrPrefix)
	if err != nil {
		return nil, err
	}

	return &FHIRClient{
		metrics:          metrics,
		client:           httpClient,
		apiAddrPrefix:    apiAddrPrefix,
		projectReference: projectID,
		locationID:       locationID,
		datasetID:        datasetID,
		fhirStoreID:      fhirStoreID,
	}, nil
}

func validatesComponents(projectID, locationID, datasetID, storeID string) error {
	if projectID == "" {
		return fmt.Errorf("missing project ID")
	}
	if locationID == "" {
		return fmt.Errorf("missing location ID")
	}
	if datasetID == "" {
		return fmt.Errorf("missing dataset ID")
	}
	if storeID == "" {
		return fmt.Errorf("missing store ID")
	}
	return nil
}

// initHTTPClient creates an HTTP client and does the authentication work.
func initHTTPClient(ctx context.Context, apiAddrPrefix string) (*http.Client, error) {
	ts, err := goauth2.DefaultTokenSource(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("oauth2google.DefaultTokenSource: %v", err)
	}

	o := []option.ClientOption{
		option.WithEndpoint(apiAddrPrefix),
		option.WithScopes(scope),
		option.WithTokenSource(ts),
	}
	log.Infof("Dialing connection to %v", apiAddrPrefix)
	httpClient, _, err := transport.NewHTTPClient(ctx, o...)
	if err != nil {
		return nil, fmt.Errorf("Dial: %v", err)
	}
	return httpClient, nil
}

// Send sends a message to the endpoint and returns the response.
// Returns an error if the request fails.
func (c *HL7Client) Send(data []byte) ([]byte, error) {
	c.metrics.Inc(sentMetric)

	msg, err := json.Marshal(sendMessageReq{Msg: message{Data: data}})
	if err != nil {
		c.metrics.Inc(sendErrorMetric)
		return nil, fmt.Errorf("failed to encode data: %v", err)
	}

	log.Infof("Sending message of size %v.", len(data))
	resp, err := c.client.Post(
		fmt.Sprintf("%v/%v/%v", c.apiAddrPrefix, util.GenerateHL7StoreName(c.projectReference, c.locationID, c.datasetID, c.hl7StoreID), sendSuffix),
		contentType, bytes.NewReader(msg))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.metrics.Inc(sendErrorMetric)
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.metrics.Inc(sendErrorMetric)
		return nil, fmt.Errorf("unable to read data from response: %v", err)
	}

	var parsedResp *sendMessageResp
	if err := json.Unmarshal(body, &parsedResp); err != nil {
		c.metrics.Inc(sendErrorMetric)
		return nil, fmt.Errorf("unable to parse response data: %v", err)
	}

	log.Infof("Message was successfully sent.")
	return parsedResp.Hl7Ack, nil
}

// Get retrieves a message from the server.
// Returns an error if the request fails.
func (c *HL7Client) Get(msgName string) ([]byte, error) {
	c.metrics.Inc(fetchedMetric)
	projectReference, locationID, datasetID, hl7StoreID, _, err := util.ParseHL7MessageName(msgName)
	if err != nil {
		c.metrics.Inc(fetchErrorInternalMetric)
		return nil, fmt.Errorf("parsing message name: %v", err)
	}
	if projectReference != c.projectReference {
		c.metrics.Inc(fetchErrorInternalMetric)
		return nil, fmt.Errorf("message name %v is not from expected project %v", msgName, c.projectReference)
	}
	if locationID != c.locationID {
		c.metrics.Inc(fetchErrorInternalMetric)
		return nil, fmt.Errorf("message name %v is not from expected location %v", msgName, c.locationID)
	}
	if datasetID != c.datasetID {
		c.metrics.Inc(fetchErrorInternalMetric)
		return nil, fmt.Errorf("message name %v is not from expected dataset %v", msgName, c.datasetID)
	}
	if hl7StoreID != c.hl7StoreID {
		c.metrics.Inc(fetchErrorInternalMetric)
		return nil, fmt.Errorf("message name %v is not from expected HL7 store %v", msgName, c.hl7StoreID)
	}

	log.Infof("Started to fetch message.")
	resp, err := c.client.Get(fmt.Sprintf("%v/%v", c.apiAddrPrefix, msgName))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.metrics.Inc(fetchErrorMetric)
		return nil, fmt.Errorf("failed to fetch message: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.metrics.Inc(sendErrorMetric)
		return nil, fmt.Errorf("unable to read data from response: %v", err)
	}
	var msg *message
	if err := json.Unmarshal(body, &msg); err != nil {
		c.metrics.Inc(sendErrorMetric)
		return nil, fmt.Errorf("unable to parse data: %v", err)
	}
	log.Infof("Message was successfully fetched.")
	return msg.Data, nil
}

// ExecuteBundle calls the FHIR transaction API with a bundle of operations.
// The response contains processing result for each entry in the bundle.
func (c *FHIRClient) ExecuteBundle(bundle []byte) ([]byte, error) {
	log.Infof("Executing bundle...")

	u := fmt.Sprintf("%v/%v", c.apiAddrPrefix, util.GenerateFHIRStoreName(c.projectReference, c.locationID, c.datasetID, c.fhirStoreID))
	resp, err := c.client.Post(u, fhirJSONContentType, bytes.NewReader(bundle))
	if err != nil {
		return []byte{}, fmt.Errorf("failed to execute bundle: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("unexpected HTTP status: %v", resp.StatusCode)
	}

	return body, nil
}
