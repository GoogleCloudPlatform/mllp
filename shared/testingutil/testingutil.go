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

// Package testingutil provides utility functions used only in tests.
package testingutil

import (
	"sync"
	"testing"
)

// CheckMetrics checks whether metrics match expected.
func CheckMetrics(t *testing.T, metrics *FakeMonitoringClient, expected map[string]int64) {
	for m, v := range expected {
		if metrics.CounterValue(m) != v {
			t.Errorf("Metric %v expected %v, got %v", m, v, metrics.CounterValue(m))
		}
	}
}

// FakeMonitoringClient is a client used for local testing. Not thread-safe.
type FakeMonitoringClient struct {
	latencies map[string][]float64
	counters  map[string]int64

	mu sync.RWMutex
}

// NewFakeMonitoringClient creates a new FakeMonitoringClient.
func NewFakeMonitoringClient() *FakeMonitoringClient {
	return &FakeMonitoringClient{latencies: make(map[string][]float64), counters: make(map[string]int64)}
}

// CounterValue returns the value of a counter metric.
func (c *FakeMonitoringClient) CounterValue(name string) int64 {
	return c.counters[name]
}

// IncCounter increment a counter metric.
func (c *FakeMonitoringClient) IncCounter(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters[name]++
}

// NewCounter creates a new counter metric.
func (c *FakeMonitoringClient) NewCounter(name, desc string) {
	c.counters[name] = 0
}

// AddLatency adds a latency value to a latency metric.
func (c *FakeMonitoringClient) AddLatency(name string, value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.latencies[name] = append(c.latencies[name], value)
}

// NewLatency creates a new latency metric.
func (c *FakeMonitoringClient) NewLatency(name, desc string) {
	c.latencies[name] = nil
}
