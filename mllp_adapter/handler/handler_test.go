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

	"shared/monitoring"
	"shared/testingutil"
)

const msgName = "projects/1/datasets/2/hl7/messagestore/messages/3"

var msgBytes = []byte("messagebody")

// The only noticeable difference between this and pubsub.Message is that
// Ack doesn't do anything.
//
// Conforms to //third_party/cloud/healthcare/hl7/shared/pubsub/pubsub.Message.
type fakeMessage struct {
	data       []byte
	attributes map[string]string

	acked bool
}

func (fm *fakeMessage) Ack() { fm.acked = true }

func (fm *fakeMessage) Data() []byte { return fm.data }

func (fm *fakeMessage) Attributes() map[string]string { return fm.attributes }

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
		sentMsgExpected []byte
		ackExpected     bool
		expectedMetrics map[string]int64
	}{
		{
			name:            "ok",
			msg:             &fakeMessage{data: []byte(msgName), attributes: map[string]string{"publish": "true"}},
			sender:          &fakeSender{},
			sentMsgExpected: msgBytes,
			ackExpected:     true,
			expectedMetrics: map[string]int64{processedMetric: 1, fetchErrorMetric: 0, sendErrorMetric: 0, ignoredMetric: 0},
		},
		{
			name:            "not published",
			msg:             &fakeMessage{data: []byte(msgName)},
			sender:          &fakeSender{},
			expectedMetrics: map[string]int64{processedMetric: 1, fetchErrorMetric: 0, sendErrorMetric: 0, ignoredMetric: 1},
		},
		{
			name:            "msg not found",
			msg:             &fakeMessage{data: []byte("invalid_name"), attributes: map[string]string{"publish": "true"}},
			sender:          &fakeSender{},
			expectedMetrics: map[string]int64{processedMetric: 1, fetchErrorMetric: 1, sendErrorMetric: 0, ignoredMetric: 0},
		},
		{
			name:            "send error",
			msg:             &fakeMessage{data: []byte(msgName), attributes: map[string]string{"publish": "true"}},
			sender:          &fakeSender{error: true},
			expectedMetrics: map[string]int64{processedMetric: 1, fetchErrorMetric: 0, sendErrorMetric: 1, ignoredMetric: 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mt := monitoring.NewClient()
			fetcher := &fakeFetcher{msgs: map[string][]byte{msgName: msgBytes}}
			handler := New(mt, fetcher, tc.sender)
			handler.Handle(tc.msg)

			if !bytes.Equal(tc.sender.msgSent, tc.sentMsgExpected) {
				t.Errorf("Expected sent message %v, got %v", tc.sentMsgExpected, tc.sender.msgSent)
			}
			if tc.msg.acked != tc.ackExpected {
				t.Errorf("Expected ack status %v, got %v", tc.ackExpected, tc.msg.acked)
			}
			testingutil.CheckMetrics(t, mt, tc.expectedMetrics)
		})
	}
}
