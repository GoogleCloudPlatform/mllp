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

package healthapiclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"shared/monitoring"
	"shared/testingutil"
	"shared/util"
)

const (
	projectReference    = "123"
	locationID          = "test-central1"
	datasetID           = "456"
	hl7StoreID          = "678"
	msgID               = "890"
	fhirStoreID         = "101"
	fhirResID           = "p1"
	pathPrefix          = "/projects/123/locations/test-central1/datasets/456/hl7Stores/678/messages/"
	sendPath            = "/projects/123/locations/test-central1/datasets/456/hl7Stores/678/messages:ingest"
	getPath             = "/projects/123/locations/test-central1/datasets/456/hl7Stores/678/messages/890"
	executeBundlePath   = "/projects/123/locations/test-central1/datasets/456/fhirStores/101"
	createdResourceName = "projects/123/locations/test-central1/datasets/456/fhirStores/101/resources/Patient/06406a39-f4ff-43e1-8ea3-6d9f542870c8"
)

var (
	received        = [][]byte{}
	toSend          = map[string][]byte{}
	cannedMsg       = []byte("abcd")
	cannedAck       = []byte("ack")
	writtenResource = ""

	patientBundle = `{
  "type":"TRANSACTION",
  "entry":[
    {
      "request":{
        "method":"PUT"
      },
      "resource":{
        "birthDate": "1975-01-01",
        "language": "abc",
        "resourceType": "Patient"
      }
    },
  ],
  "resourceType":"Bundle"
}`
	executeBundleResp = fmt.Sprintf(`{
  "type":"TRANSACTION_RESPONSE",
  "entry":[
    {
      "response":{
        "location": "%v",
      }
    },
  ],
  "resourceType":"Bundle"
}`, createdResourceName)
)

func setUp() *httptest.Server {
	received = [][]byte{}
	toSend = map[string][]byte{}

	return httptest.NewTLSServer(http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			switch url := req.URL.EscapedPath(); url {
			case sendPath:
				defer req.Body.Close()
				b, err := ioutil.ReadAll(req.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				var msgReq *sendMessageReq
				if err := json.Unmarshal(b, &msgReq); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				received = append(received, msgReq.Msg.Data)

				data, err := json.Marshal(&sendMessageResp{Hl7Ack: cannedAck})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Write(data)
			case getPath:
				id := strings.TrimPrefix(url, pathPrefix)
				msg, ok := toSend[id]
				if !ok {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				data, err := json.Marshal(&message{Data: msg})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Write(data)
			case executeBundlePath:
				defer req.Body.Close()
				body, err := ioutil.ReadAll(req.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				writtenResource = string(body)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(executeBundleResp))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		},
	))
}

// newHL7Client creates a new Client pointed to a fake HL7 service.
func newHL7Client(client *http.Client, apiAddrPrefix, projectReference, locationID, datasetID, hl7StoreID string) *HL7Client {
	c := &HL7Client{
		metrics:          monitoring.NewClient(),
		client:           client,
		apiAddrPrefix:    apiAddrPrefix,
		projectReference: projectReference,
		locationID:       locationID,
		datasetID:        datasetID,
		hl7StoreID:       hl7StoreID,
	}
	c.initMetrics()
	return c
}

// newFHIRClient creates a new Client pointed to a fake FHIR service.
func newFHIRClient(client *http.Client, apiAddrPrefix, projectReference, locationID, datasetID, fhirStoreID string) *FHIRClient {
	return &FHIRClient{
		metrics:          monitoring.NewClient(),
		client:           client,
		apiAddrPrefix:    apiAddrPrefix,
		projectReference: projectReference,
		locationID:       locationID,
		datasetID:        datasetID,
		fhirStoreID:      fhirStoreID,
	}
}

func TestSend(t *testing.T) {
	testCases := []struct {
		name             string
		projectReference string
		datasetID        string
		hl7StoreID       string
		msgs             [][]byte
		expectedMetrics  map[string]int64
	}{
		{
			"single message",
			projectReference,
			datasetID,
			hl7StoreID,
			[][]byte{cannedMsg},
			map[string]int64{sentMetric: 1, sendErrorMetric: 0},
		},
		{
			"multiple messages",
			projectReference,
			datasetID,
			hl7StoreID,
			[][]byte{cannedMsg, cannedMsg, cannedMsg},
			map[string]int64{sentMetric: 3, sendErrorMetric: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := setUp()
			defer s.Close()
			c := newHL7Client(s.Client(), s.URL, tc.projectReference, locationID, tc.datasetID, tc.hl7StoreID)
			for _, msg := range tc.msgs {
				ack, err := c.Send(msg)
				if err != nil {
					t.Errorf("Unexpected send error: %v", err)
				}
				if !reflect.DeepEqual(cannedAck, ack) {
					t.Errorf("Expected ack %v but got %v", cannedAck, ack)
				}
			}
			if !reflect.DeepEqual(tc.msgs, received) {
				t.Errorf("Messages differ: expected %v but got %v", tc.msgs, received)
			}
			testingutil.CheckMetrics(t, c.metrics, tc.expectedMetrics)
		})
	}
}

func TestSendError(t *testing.T) {
	testCases := []struct {
		name             string
		projectReference string
		datasetID        string
		hl7StoreID       string
		msgs             [][]byte
		expectedMetrics  map[string]int64
	}{
		{
			"wrong project ID",
			"wrongproject",
			datasetID,
			hl7StoreID,
			[][]byte{cannedMsg},
			map[string]int64{sentMetric: 1, sendErrorMetric: 1},
		},
		{
			"wrong dataset ID",
			projectReference,
			"wrongdataset",
			hl7StoreID,
			[][]byte{cannedMsg},
			map[string]int64{sentMetric: 1, sendErrorMetric: 1},
		},
		{
			"wrong HL7 store ID",
			projectReference,
			datasetID,
			"wronghl7store",
			[][]byte{cannedMsg},
			map[string]int64{sentMetric: 1, sendErrorMetric: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := setUp()
			defer s.Close()
			c := newHL7Client(s.Client(), s.URL, tc.projectReference, locationID, tc.datasetID, tc.hl7StoreID)
			for _, msg := range tc.msgs {
				ack, err := c.Send(msg)
				if err == nil {
					t.Errorf("Expected send error but got %v", ack)
				}
			}
			if len(received) != 0 {
				t.Errorf("Unexpected messages received: %v", received)
			}
			testingutil.CheckMetrics(t, c.metrics, tc.expectedMetrics)
		})
	}
}

func TestGet(t *testing.T) {
	s := setUp()
	defer s.Close()
	toSend = map[string][]byte{msgID: cannedMsg}
	c := newHL7Client(s.Client(), s.URL, projectReference, locationID, datasetID, hl7StoreID)
	msg, err := c.Get(util.GenerateHL7MessageName(projectReference, locationID, datasetID, hl7StoreID, msgID))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(cannedMsg, msg) {
		t.Errorf("Expected msg %v but got %v", cannedMsg, msg)
	}
	expectedMetrics := map[string]int64{fetchedMetric: 1, fetchErrorMetric: 0, fetchErrorInternalMetric: 0}
	testingutil.CheckMetrics(t, c.metrics, expectedMetrics)
}

func TestGetError(t *testing.T) {
	testCases := []struct {
		name            string
		msgName         string
		expectedMetrics map[string]int64
	}{
		{
			"wrong project ID",
			util.GenerateHL7MessageName("wrong", locationID, datasetID, hl7StoreID, msgID),
			map[string]int64{fetchedMetric: 1, fetchErrorMetric: 0, fetchErrorInternalMetric: 1},
		},
		{
			"wrong dataset ID",
			util.GenerateHL7MessageName(projectReference, locationID, "wrong", hl7StoreID, msgID),
			map[string]int64{fetchedMetric: 1, fetchErrorMetric: 0, fetchErrorInternalMetric: 1},
		},
		{
			"wrong HL7 store ID",
			util.GenerateHL7MessageName(projectReference, locationID, datasetID, "wrong", msgID),
			map[string]int64{fetchedMetric: 1, fetchErrorMetric: 0, fetchErrorInternalMetric: 1},
		},
		{
			"cannot be parsed",
			"garbage",
			map[string]int64{fetchedMetric: 1, fetchErrorMetric: 0, fetchErrorInternalMetric: 1},
		},
		{
			"message not found",
			util.GenerateHL7MessageName(projectReference, locationID, datasetID, hl7StoreID, "wrong"),
			map[string]int64{fetchedMetric: 1, fetchErrorMetric: 1, fetchErrorInternalMetric: 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := setUp()
			defer s.Close()
			c := newHL7Client(s.Client(), s.URL, projectReference, locationID, datasetID, hl7StoreID)
			msg, err := c.Get(tc.msgName)
			if err == nil {
				t.Errorf("Expected error but got %v", msg)
			}
			testingutil.CheckMetrics(t, c.metrics, tc.expectedMetrics)
		})
	}
}

func TestExecuteBundle(t *testing.T) {
	s := setUp()
	defer s.Close()
	c := newFHIRClient(s.Client(), s.URL, projectReference, locationID, datasetID, fhirStoreID)
	_, err := c.ExecuteBundle([]byte(patientBundle))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if writtenResource != patientBundle {
		t.Errorf("Server didn't receive the same bundle, got: \n%v\n, wanted: \n%v\n", writtenResource, patientBundle)
	}
}

func TestExecuteBundle_Errors(t *testing.T) {
	testCases := []struct {
		name             string
		projectReference string
		locationID       string
		datasetID        string
		fhirStoreID      string
	}{
		{
			"wrong project ID",
			"wrong",
			locationID,
			datasetID,
			fhirStoreID,
		},
		{
			"wrong dataset ID",
			projectReference,
			locationID,
			"wrong",
			fhirStoreID,
		},
		{
			"wrong location ID",
			projectReference,
			"wrong",
			datasetID,
			fhirStoreID,
		},
		{
			"wrong FHIR store ID",
			projectReference,
			locationID,
			datasetID,
			"wrong",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := setUp()
			defer s.Close()
			c := newFHIRClient(s.Client(), s.URL, tc.projectReference, tc.locationID, tc.datasetID, tc.fhirStoreID)
			_, err := c.ExecuteBundle([]byte(patientBundle))
			if err == nil {
				t.Errorf("Expected error but got none")
			}
		})
	}
}
