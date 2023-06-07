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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	log "github.com/golang/glog"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"github.com/GoogleCloudPlatform/mllp/shared/monitoring"
	"github.com/GoogleCloudPlatform/mllp/shared/util"

	healthcare "google.golang.org/api/healthcare/v1"
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

// HL7V2Client represents a client of the HL7v2 API.
type HL7V2Client struct {
	metrics      monitoring.Client
	storeService *healthcare.ProjectsLocationsDatasetsHl7V2StoresService
	projectID    string
	locationID   string
	datasetID    string
	hl7V2StoreID string
	logNACKedMsg bool
	logErrMsg    bool
}

type sendMessageErrorResp struct {
	Error struct {
		Message string `json:"message"`
		Details []struct {
			Hl7Nack []byte `json:"hl7Nack"`
		} `json:"details"`
	} `json:"error"`
}

// StoreInfo contains information about an HL7v2 store.
type StoreInfo struct {
	ProjectID    string
	LocationID   string
	DatasetID    string
	HL7V2StoreID string
}

// Option contains flags for the HL7v2Client.
type Option struct {
	LogNACKedMessage bool
	LogErrorMessage  bool
}

// NewHL7V2Client creates a properly authenticated client that talks to an HL7v2 backend.
func NewHL7V2Client(ctx context.Context, cred string, metrics monitoring.Client, si StoreInfo, opt Option) (*HL7V2Client, error) {
	if err := validatesComponents(si.ProjectID, si.LocationID, si.DatasetID, si.HL7V2StoreID); err != nil {
		return nil, err
	}

	storeService, err := initHL7v2StoreService(ctx, cred)
	if err != nil {
		return nil, err
	}

	c := &HL7V2Client{
		metrics:      metrics,
		storeService: storeService,
		projectID:    si.ProjectID,
		locationID:   si.LocationID,
		datasetID:    si.DatasetID,
		hl7V2StoreID: si.HL7V2StoreID,
		logNACKedMsg: opt.LogNACKedMessage,
		logErrMsg:    opt.LogErrorMessage,
	}
	c.initMetrics()
	return c, nil
}

func (c *HL7V2Client) initMetrics() {
	c.metrics.NewCounter(sentMetric, "Number of HL7 messages sent to HL7 store.")
	c.metrics.NewCounter(sendErrorMetric, "Number of errors when sending HL7 message to HL7 store.")
	c.metrics.NewCounter(fetchedMetric, "Number of HL7 messages fetched from HL7 Store.")
	c.metrics.NewCounter(fetchErrorMetric, "Number of errors when fetching HL7 message from HL7 Store.")
	c.metrics.NewCounter(fetchErrorInternalMetric, "Number of adapter internal errors when fetching HL7 message from HL7 Store.")
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

// initHL7v2StoreService creates an HL7v2 store service and does the
// authentication work.
func initHL7v2StoreService(ctx context.Context, cred string) (*healthcare.ProjectsLocationsDatasetsHl7V2StoresService, error) {
	ts, err := util.TokenSource(ctx, cred, scope)
	if err != nil {
		return nil, fmt.Errorf("oauth2google.DefaultTokenSource: %v", err)
	}

	healthcareService, err := healthcare.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("healthcare.NewService: %v", err)
	}
	return healthcareService.Projects.Locations.Datasets.Hl7V2Stores, nil
}

// Send sends a message to the endpoint and returns the ACK/NACK response.
// Returns an error if the request fails without a NACK response.
func (c *HL7V2Client) Send(data []byte) ([]byte, error) {
	c.metrics.IncCounter(sentMetric)

	req := &healthcare.IngestMessageRequest{
		Message: &healthcare.Message{
			Data: base64.StdEncoding.EncodeToString(data),
		},
	}
	ctx := context.Background()
	log.Infof("Received message of size %v bytes. Sending this message to the Cloud Healthcare API HL7V2 Store.", len(data))
	parent := util.GenerateHL7V2StoreName(c.projectID, c.locationID, c.datasetID, c.hl7V2StoreID)
	ingest := c.storeService.Messages.Ingest(parent, req)
	ingest.Header().Add("X-GOOG-API-FORMAT-VERSION", "2")
	resp, err := ingest.Context(ctx).Do()
	if err != nil {
		c.metrics.IncCounter(sendErrorMetric)
		if e, ok := err.(*googleapi.Error); ok {
			if len(e.Body) == 0 {
				return nil, e
			}
			nack, em, err := extractNACKAndErrorMessage([]byte(e.Body))
			if err != nil {
				return nil, err
			}
			if nack != nil {
				log.Errorf("Message was sent to the Cloud Healthcare API HL7V2 Store, received a NACK response.")
				if c.logNACKedMsg {
					log.Errorf("The original message was %s", sanitizeMessageForPrintout(data))
				}
				if c.logErrMsg {
					log.Errorf("The error message was %q", em)
				}
				return nack, nil
			}
		}
		return nil, err
	}

	ack, err := base64.StdEncoding.DecodeString(resp.Hl7Ack)
	if err != nil {
		c.metrics.IncCounter(sendErrorMetric)
		return nil, fmt.Errorf("unable to parse ACK response: %v", err)
	}
	log.Infof("Message was successfully sent to the Cloud Healthcare API HL7V2 Store.")
	return ack, nil
}

func sanitizeMessageForPrintout(data []byte) string {
	if utf8.Valid(data) {
		// Convert to UTF8 if possible.
		return string(data)
	}
	// Otherwise output base64 so that users can at least get the full data.
	return fmt.Sprintf("[base64 encoded] %s", base64.StdEncoding.EncodeToString(data))
}

func extractNACKAndErrorMessage(resp []byte) ([]byte, string, error) {
	var parsedErrorResp *sendMessageErrorResp
	if err := json.Unmarshal(resp, &parsedErrorResp); err != nil {
		return nil, "", fmt.Errorf("unable to parse error response data: %v\n%s", err, resp)
	}
	em := parsedErrorResp.Error.Message
	for _, d := range parsedErrorResp.Error.Details {
		if len(d.Hl7Nack) > 0 {
			return d.Hl7Nack, em, nil
		}
	}
	return nil, em, nil
}

// Get retrieves a message from the server.
// Returns an error if the request fails.
func (c *HL7V2Client) Get(msgName string) ([]byte, error) {
	c.metrics.IncCounter(fetchedMetric)
	projectID, locationID, datasetID, hl7V2StoreID, _, err := util.ParseHL7V2MessageName(msgName)
	if err != nil {
		c.metrics.IncCounter(fetchErrorInternalMetric)
		return nil, fmt.Errorf("parsing message name: %v", err)
	}
	if projectID != c.projectID {
		c.metrics.IncCounter(fetchErrorInternalMetric)
		return nil, fmt.Errorf("message name %v is not from expected project %v", msgName, c.projectID)
	}
	if locationID != c.locationID {
		c.metrics.IncCounter(fetchErrorInternalMetric)
		return nil, fmt.Errorf("message name %v is not from expected location %v", msgName, c.locationID)
	}
	if datasetID != c.datasetID {
		c.metrics.IncCounter(fetchErrorInternalMetric)
		return nil, fmt.Errorf("message name %v is not from expected dataset %v", msgName, c.datasetID)
	}
	if hl7V2StoreID != c.hl7V2StoreID {
		c.metrics.IncCounter(fetchErrorInternalMetric)
		return nil, fmt.Errorf("message name %v is not from expected HL7v2 store %v", msgName, c.hl7V2StoreID)
	}

	log.Infof("Started to fetch message from the Cloud Healthcare API HL7V2 Store")
	resp, err := c.storeService.Messages.Get(msgName).Context(context.Background()).Do()
	if err != nil {
		c.metrics.IncCounter(fetchErrorMetric)
		return nil, fmt.Errorf("failed to fetch message: %v", err)
	}
	msg, err := base64.StdEncoding.DecodeString(resp.Data)
	if err != nil {
		c.metrics.IncCounter(fetchErrorMetric)
		return nil, fmt.Errorf("unable to parse data: %v", err)
	}
	log.Infof("Message was successfully fetched from the Cloud Healthcare API HL7V2 Store.")
	return msg, nil
}
