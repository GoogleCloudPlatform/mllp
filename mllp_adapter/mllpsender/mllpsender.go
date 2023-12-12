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

// Package mllpsender sends HL7 messages via MLLP.
package mllpsender

import (
	"sync"
	"fmt"
	"net"

	"github.com/GoogleCloudPlatform/mllp/mllp_adapter/mllp"
	"github.com/GoogleCloudPlatform/mllp/shared/monitoring"
)

const (
	sentMetric      = "mllpsender-messages-sent"
	ackErrorMetric  = "mllpsender-messages-ack-error"
	sendErrorMetric = "mllpsender-messages-send-error"
	dialErrorMetric = "mllpsender-connections-dial-error"
	poolErrorMetric = "mllpsender-connections-pool-error"
)

var conn net.Conn

// MLLPSender represents an MLLP sender.
type MLLPSender struct {
	addr    string
	metrics monitoring.Client
	pool *sync.Pool
}

// NewSender creates a new MLLPSender.
func NewSender(addr string, metrics monitoring.Client) *MLLPSender {
	metrics.NewCounter(sentMetric, "Number of HL7 messages sent to mllp_addr")
	metrics.NewCounter(ackErrorMetric, "Number of errors when receiving ACK from mllp_addr")
	metrics.NewCounter(sendErrorMetric, "Number of errors when sending HL7 message to mllp_addr")
	metrics.NewCounter(dialErrorMetric, "Number of errors when dialing to mllp_addr")
	metrics.NewCounter(poolErrorMetric, "Number of errors when getting connection from pool")

	pool := &sync.Pool{
		New: func() interface{} {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				panic(fmt.Errorf("dialing: %w", err))
			}
			return conn
		},
	}

	return &MLLPSender{addr: addr, metrics: metrics, pool: pool}
}

// Send sends an HL7 messages via MLLP.
func (m *MLLPSender) Send(msg []byte) ([]byte, error) {
	m.metrics.IncCounter(sentMetric)
	
	conn, ok := m.pool.Get().(net.Conn)
	if !ok {
		m.metrics.IncCounter(poolErrorMetric)
		return nil, fmt.Errorf("unexpected object type returned by pool: %T", conn)
	  }
	if err := mllp.WriteMsg(conn, msg); err != nil {
		m.metrics.IncCounter(sendErrorMetric)
		return nil, fmt.Errorf("writing message: %v", err)
	}
	ack, err := mllp.ReadMsg(conn)
	if err != nil {
		m.metrics.IncCounter(ackErrorMetric)
		return nil, fmt.Errorf("reading ACK: %v", err)
	}
	m.pool.Put(conn)
	return ack, nil
}
