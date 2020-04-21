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

// Package pubsub handles notifications of messages that should be sent back to
// the partner.
package pubsub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
	"github.com/GoogleCloudPlatform/mllp/shared/util"
)

const (
	fetchErrorMetric = "pubsub-messages-fetch-error"
	sendErrorMetric  = "pubsub-messages-send-error"
	processedMetric  = "pubsub-messages-processed"
	ignoredMetric    = "pubsub-messages-ignored"

	scope = "https://www.googleapis.com/auth/pubsub"
)

// Message represents a pubsub message.
type Message interface {
	Ack()
	Data() []byte
	Attrs() map[string]string
}

type messageWrapper struct {
	msg *pubsub.Message
}

func (m *messageWrapper) Ack() {
	m.msg.Ack()
}

func (m *messageWrapper) Data() []byte {
	return m.msg.Data
}

func (m *messageWrapper) Attrs() map[string]string {
	return m.msg.Attributes
}

// MessageHandler is the interface for handling HL7v2 messages.
type MessageHandler interface {
	Handle(Message)
}

// Listen listens for notifications from a pubsub subscription, uses the ids
// in the messages to fetch content with the HL7v2 API, then sends the message
// to the partner over MLLP.
func Listen(ctx context.Context, cred string, h MessageHandler, projectID string, topic string, opts ...option.ClientOption) error {
	ts, err := util.TokenSource(ctx, cred, scope)
	if err != nil {
		return fmt.Errorf("getting default token source: %v", err)
	}
	fullOpts := []option.ClientOption{option.WithTokenSource(ts)}
	fullOpts = append(fullOpts, opts...)
	return listen(ctx, h, projectID, topic, fullOpts...)
}

// listen omits the TokenSource code for test purposes.
func listen(ctx context.Context, h MessageHandler, projectID string, topic string, opts ...option.ClientOption) error {
	client, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		return fmt.Errorf("creating pubsub client: %v", err)
	}

	return client.Subscription(topic).Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		h.Handle(&messageWrapper{msg: msg})
	})
}
