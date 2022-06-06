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

package handler

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/mllp/shared/testingutil"
)

const (
	msgName = "projects/1/datasets/2/hl7/messagestore/messages/3"
)

var (
	msgBytes = []byte("messagebody")
)

type fakeMessage struct {
	name    string
	acked   bool
	publish bool
}

func (m *fakeMessage) Ack() {
	m.acked = true
}

func (m *fakeMessage) Data() []byte {
	return []byte(m.name)
}

func (m *fakeMessage) Attrs() map[string]string {
	if !m.publish {
		return map[string]string{}
	}
	return map[string]string{"publish": "true"}
}

type fakeFetcher struct {
	msgs map[string][]byte
}

func (f *fakeFetcher) Get(name string) ([]byte, error) {
	msg, ok := f.msgs[name]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return msg, nil
}

type fakeSender struct {
	error   bool
	msgSent []byte
}

func (s *fakeSender) Send(msg []byte) ([]byte, error) {
	if s.error {
		return nil, fmt.Errorf("send error")
	}
	s.msgSent = msg
	return nil, nil
}

func TestHandle(t *testing.T) {
	testCases := []struct {
		name            string
		msg             *fakeMessage
		sender          *fakeSender
		checkPublish    bool
		sentMsgExpected []byte
		ackExpected     bool
		expectedMetrics map[string]int64
	}{
		{
			name:            "ok",
			msg:             &fakeMessage{name: msgName, publish: true},
			sender:          &fakeSender{},
			checkPublish:    true,
			sentMsgExpected: msgBytes,
			ackExpected:     true,
			expectedMetrics: map[string]int64{processedMetric: 1, fetchErrorMetric: 0, sendErrorMetric: 0, ignoredMetric: 0},
		},
		{
			name:            "not published",
			msg:             &fakeMessage{name: msgName, publish: false},
			sender:          &fakeSender{},
			checkPublish:    true,
			expectedMetrics: map[string]int64{processedMetric: 1, fetchErrorMetric: 0, sendErrorMetric: 0, ignoredMetric: 1},
		},
		{
			name:            "not published - disable check",
			msg:             &fakeMessage{name: msgName, publish: false},
			sender:          &fakeSender{},
			checkPublish:    false,
			sentMsgExpected: msgBytes,
			ackExpected:     true,
			expectedMetrics: map[string]int64{processedMetric: 1, fetchErrorMetric: 0, sendErrorMetric: 0, ignoredMetric: 0},
		},
		{
			name:            "msg not found",
			msg:             &fakeMessage{name: "invalid_name", publish: true},
			sender:          &fakeSender{},
			checkPublish:    true,
			expectedMetrics: map[string]int64{processedMetric: 1, fetchErrorMetric: 1, sendErrorMetric: 0, ignoredMetric: 0},
		},
		{
			name:            "send error",
			msg:             &fakeMessage{name: msgName, publish: true},
			sender:          &fakeSender{error: true},
			checkPublish:    true,
			expectedMetrics: map[string]int64{processedMetric: 1, fetchErrorMetric: 0, sendErrorMetric: 1, ignoredMetric: 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fc := testingutil.NewFakeMonitoringClient()
			fetcher := &fakeFetcher{msgs: map[string][]byte{msgName: msgBytes}}
			handler := New(fc, fetcher, tc.sender, tc.checkPublish)
			handler.Handle(tc.msg)

			if !bytes.Equal(tc.sender.msgSent, tc.sentMsgExpected) {
				t.Errorf("Expected sent message %v, got %v", tc.sentMsgExpected, tc.sender.msgSent)
			}
			if tc.msg.acked != tc.ackExpected {
				t.Errorf("Expected ack status %v, got %v", tc.ackExpected, tc.msg.acked)
			}
			testingutil.CheckMetrics(t, fc, tc.expectedMetrics)
		})
	}
}
