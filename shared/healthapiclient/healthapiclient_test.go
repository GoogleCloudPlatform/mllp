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
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"google.golang.org/api/option"
	"github.com/GoogleCloudPlatform/mllp/shared/testingutil"
	"github.com/GoogleCloudPlatform/mllp/shared/util"

	healthcare "google.golang.org/api/healthcare/v1"
)

const (
	projectID             = "123"
	locationID            = "test-central1"
	datasetID             = "456"
	hl7V2StoreID          = "678"
	msgID                 = "890"
	pathPrefix            = "/v1/projects/123/locations/test-central1/datasets/456/hl7V2Stores/678/messages/"
	sendPath              = "/v1/projects/123/locations/test-central1/datasets/456/hl7V2Stores/678/messages:ingest"
	getPath               = "/v1/projects/123/locations/test-central1/datasets/456/hl7V2Stores/678/messages/890"
	invalidErrResp        = "invalid error response"
	rateLimitExceededResp = "too many requests response"
)

var (
	received  = [][]byte{}
	toSend    = map[string][]byte{}
	cannedMsg = []byte("abcd")
	cannedAck = []byte("ack")
)

type message struct {
	Data []byte `json:"data"`
}

type sendMessageReq struct {
	Msg message `json:"message"`
}

type sendMessageResp struct {
	Hl7Ack []byte `json:"hl7Ack"`
}

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

				data, err := json.Marshal(&sendMessageResp{Hl7Ack: cannedAck})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if string(msgReq.Msg.Data) == rateLimitExceededResp {
					w.WriteHeader(http.StatusTooManyRequests)
					w.Write([]byte("{\"reason\":\"RATE_LIMIT_EXCEEDED\"}"))
					return
				}

				if string(msgReq.Msg.Data) == invalidErrResp {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("{"))
					return
				}
				received = append(received, msgReq.Msg.Data)

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
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		},
	))
}

// newHL7V2Client creates a new Client pointed to a fake HL7v2 service.
func newHL7V2Client(client *http.Client, apiAddrPrefix, projectID, locationID, datasetID, hl7V2StoreID string) *HL7V2Client {
	s, _ := healthcare.NewService(context.Background(), option.WithHTTPClient(client))
	s.BasePath = apiAddrPrefix
	c := &HL7V2Client{
		metrics:      testingutil.NewFakeMonitoringClient(),
		storeService: s.Projects.Locations.Datasets.Hl7V2Stores,
		projectID:    projectID,
		locationID:   locationID,
		datasetID:    datasetID,
		hl7V2StoreID: hl7V2StoreID,
	}
	c.initMetrics()
	return c
}

func TestSend(t *testing.T) {
	testCases := []struct {
		name            string
		projectID       string
		datasetID       string
		hl7V2StoreID    string
		msgs            [][]byte
		expectedMetrics map[string]int64
	}{
		{
			"single message",
			projectID,
			datasetID,
			hl7V2StoreID,
			[][]byte{cannedMsg},
			map[string]int64{sentMetric: 1, sendErrorMetric: 0},
		},
		{
			"multiple messages",
			projectID,
			datasetID,
			hl7V2StoreID,
			[][]byte{cannedMsg, cannedMsg, cannedMsg},
			map[string]int64{sentMetric: 3, sendErrorMetric: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := setUp()
			defer s.Close()
			c := newHL7V2Client(s.Client(), s.URL, tc.projectID, locationID, tc.datasetID, tc.hl7V2StoreID)
			c.metrics = testingutil.NewFakeMonitoringClient()
			for _, msg := range tc.msgs {
				ack, err := c.Send(msg)
				if err != nil {
					t.Errorf("Unexpected send error: %v", err)
				}
				if !reflect.DeepEqual(cannedAck, ack) {
					t.Errorf("Expected ack %s but got %s", cannedAck, ack)
				}
			}
			if !reflect.DeepEqual(tc.msgs, received) {
				t.Errorf("Messages differ: expected %v but got %v", tc.msgs, received)
			}
			testingutil.CheckMetrics(t, c.metrics.(*testingutil.FakeMonitoringClient), tc.expectedMetrics)
		})
	}
}

func TestSendError(t *testing.T) {
	testCases := []struct {
		name            string
		projectID       string
		datasetID       string
		hl7V2StoreID    string
		msgs            [][]byte
		expectedError   string
		expectedMetrics map[string]int64
	}{
		{
			"wrong project ID",
			"wrongproject",
			datasetID,
			hl7V2StoreID,
			[][]byte{cannedMsg},
			"response code 404",
			map[string]int64{sentMetric: 1, sendErrorMetric: 1},
		},
		{
			"wrong dataset ID",
			projectID,
			"wrongdataset",
			hl7V2StoreID,
			[][]byte{cannedMsg},
			"response code 404",
			map[string]int64{sentMetric: 1, sendErrorMetric: 1},
		},
		{
			"wrong HL7v2 store ID",
			projectID,
			datasetID,
			"wronghl7v2store",
			[][]byte{cannedMsg},
			"response code 404",
			map[string]int64{sentMetric: 1, sendErrorMetric: 1},
		},
		{
			"invalid error response",
			projectID,
			datasetID,
			hl7V2StoreID,
			[][]byte{[]byte(invalidErrResp)},
			"unable to parse error response",
			map[string]int64{sentMetric: 1, sendErrorMetric: 1},
		},
		{
			"too many requests response",
			projectID,
			datasetID,
			hl7V2StoreID,
			[][]byte{[]byte(rateLimitExceededResp)},
			"response code 429",
			map[string]int64{sentMetric: 1, sendErrorMetric: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := setUp()
			defer s.Close()
			c := newHL7V2Client(s.Client(), s.URL, tc.projectID, locationID, tc.datasetID, tc.hl7V2StoreID)
			c.metrics = testingutil.NewFakeMonitoringClient()
			for _, msg := range tc.msgs {
				ack, err := c.Send(msg)
				if err == nil {
					t.Errorf("Expected send error but got %v", ack)
				}
				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Got send error: %v, want it to contain: %v", err, tc.expectedError)
				}
			}
			if len(received) != 0 {
				t.Errorf("Unexpected messages received: %v", received)
			}
			testingutil.CheckMetrics(t, c.metrics.(*testingutil.FakeMonitoringClient), tc.expectedMetrics)
		})
	}
}

func TestGet(t *testing.T) {
	s := setUp()
	defer s.Close()
	toSend = map[string][]byte{msgID: cannedMsg}
	c := newHL7V2Client(s.Client(), s.URL, projectID, locationID, datasetID, hl7V2StoreID)
	c.metrics = testingutil.NewFakeMonitoringClient()
	msg, err := c.Get(util.GenerateHL7V2MessageName(projectID, locationID, datasetID, hl7V2StoreID, msgID))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(cannedMsg, msg) {
		t.Errorf("Expected msg %v but got %v", cannedMsg, msg)
	}
	expectedMetrics := map[string]int64{fetchedMetric: 1, fetchErrorMetric: 0, fetchErrorInternalMetric: 0}
	testingutil.CheckMetrics(t, c.metrics.(*testingutil.FakeMonitoringClient), expectedMetrics)
}

func TestGetError(t *testing.T) {
	testCases := []struct {
		name            string
		msgName         string
		expectedMetrics map[string]int64
	}{
		{
			"wrong project ID",
			util.GenerateHL7V2MessageName("wrong", locationID, datasetID, hl7V2StoreID, msgID),
			map[string]int64{fetchedMetric: 1, fetchErrorMetric: 0, fetchErrorInternalMetric: 1},
		},
		{
			"wrong dataset ID",
			util.GenerateHL7V2MessageName(projectID, locationID, "wrong", hl7V2StoreID, msgID),
			map[string]int64{fetchedMetric: 1, fetchErrorMetric: 0, fetchErrorInternalMetric: 1},
		},
		{
			"wrong HL7v2 store ID",
			util.GenerateHL7V2MessageName(projectID, locationID, datasetID, "wrong", msgID),
			map[string]int64{fetchedMetric: 1, fetchErrorMetric: 0, fetchErrorInternalMetric: 1},
		},
		{
			"cannot be parsed",
			"garbage",
			map[string]int64{fetchedMetric: 1, fetchErrorMetric: 0, fetchErrorInternalMetric: 1},
		},
		{
			"message not found",
			util.GenerateHL7V2MessageName(projectID, locationID, datasetID, hl7V2StoreID, "wrong"),
			map[string]int64{fetchedMetric: 1, fetchErrorMetric: 1, fetchErrorInternalMetric: 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := setUp()
			defer s.Close()
			c := newHL7V2Client(s.Client(), s.URL, projectID, locationID, datasetID, hl7V2StoreID)
			c.metrics = testingutil.NewFakeMonitoringClient()
			msg, err := c.Get(tc.msgName)
			if err == nil {
				t.Errorf("Expected error but got %v", msg)
			}
			testingutil.CheckMetrics(t, c.metrics.(*testingutil.FakeMonitoringClient), tc.expectedMetrics)
		})
	}
}

func TestSanitizeMessageForPrintout(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "utf8",
			data: []byte("abc"),
			want: "abc",
		},
		{
			name: "base64",
			data: []byte{0xbd, 0xb2, 0x3d},
			want: "[base64 encoded] vbI9",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeMessageForPrintout(tc.data)
			if got != tc.want {
				t.Errorf("sanitizeMessageForPrintout returned %s; want %s", got, tc.want)
			}
		})
	}
}
