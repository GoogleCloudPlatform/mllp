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

// Package monitoring implements the functionality to export timeseries data to
// the Cloud Monitoring service.
package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/golang/glog"
	"cloud.google.com/go/compute/metadata"
	"google.golang.org/api/option"
	"github.com/google/uuid"
	"github.com/GoogleCloudPlatform/mllp/shared/util"

	timestamppb "github.com/golang/protobuf/ptypes/timestamp"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"contrib.go.opencensus.io/exporter/stackdriver"
)

const scope = "https://www.googleapis.com/auth/monitoring.write"
const metricPrefix = "custom.googleapis.com/cloud/healthcare/mllp/"

// Client interface provides basic functionality to create, record and retrieve metric values
type Client interface {
	IncCounter(name string)
	NewCounter(name, desc string)
	AddLatency(name string, value float64)
	NewLatency(name, desc string)
}

// NewExportingClient returns a client that can export to metrics to Cloud Monitoring.
func NewExportingClient() *ExportingClient {
	return &ExportingClient{
		labels:    &stackdriver.Labels{},
		counters:  make(map[string]*stats.Int64Measure),
		latencies: make(map[string]*stats.Float64Measure)}
}

// ExportingClient represents a client that exports to Cloud Monitoring
type ExportingClient struct {
	projectID string
	labels    *stackdriver.Labels
	exporter  *stackdriver.Exporter

	// mu guards metrics.  The other fields are immutable
	mu        sync.RWMutex
	counters  map[string]*stats.Int64Measure
	latencies map[string]*stats.Float64Measure
}

// IncCounter increases a counter metric or does nothing if the client is nil.
func (m *ExportingClient) IncCounter(name string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	ctx := context.Background()
	stats.Record(ctx, m.counters[name].M(1))
}

// NewCounter creates a new counter metrics or does nothing if the client is nil.
func (m *ExportingClient) NewCounter(name, description string) {
	if m == nil {
		return
	}
	m.counters[name] = stats.Int64(name, description, stats.UnitDimensionless)
	v := &view.View{
		Name:        metricPrefix + name,
		Measure:     m.counters[name],
		Aggregation: view.Count(),
	}
	if err := view.Register(v); err != nil {
		log.Errorf("Failed to register the view: %v", err)
	}
}

// AddLatency adds a latency metric or does nothing if the client is nil.
func (m *ExportingClient) AddLatency(name string, value float64) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	ctx := context.Background()
	stats.Record(ctx, m.latencies[name].M(value))
}

// NewLatency creates a new latency metrics or does nothing if the client is nil.
func (m *ExportingClient) NewLatency(name, description string) {
	if m == nil {
		return
	}
	m.latencies[name] = stats.Float64(name, description, stats.UnitMilliseconds)
	v := &view.View{
		Name:    metricPrefix + name,
		Measure: m.latencies[name],
		// Latency in buckets:
		// [>=0ms, >=50ms, >=100ms, >=200ms, >=400ms, >=1s, >=2s, >=4s]
		Aggregation: view.Distribution(0, 50, 100, 200, 400, 1000, 2000, 4000),
	}
	if err := view.Register(v); err != nil {
		log.Errorf("Failed to register the view: %v", err)
	}
}

// StartExport metrics to the monitoring service roughly once a minute.
// It fetches metadata about the GCP environment and fails if not
// running on GCE or GKE.
func (m *ExportingClient) StartExport(ctx context.Context, cred string) error {
	if !metadata.OnGCE() {
		return fmt.Errorf("not running on GCE - metrics cannot be exported")
	}
	var err error
	ts, err := util.TokenSource(ctx, cred, scope)
	if err != nil {
		return fmt.Errorf("getting default token source: %v", err)
	}
	m.projectID, err = metadata.ProjectID()
	if err != nil {
		return err
	}
	zone, err := metadata.Zone()
	if err != nil {
		return err
	}
	m.labels.Set("zone", zone, "")

	instance, err := metadata.InstanceID()
	if err != nil {
		return err
	}
	m.labels.Set("instance", instance, "")
	m.labels.Set("job", "mllp_adapter", "")
	// Use uuid to prevent multiple mllp adapter instances from submitting
	// metrics with the same label.
	id := uuid.New().String()
	m.labels.Set("id", id, "")
	log.Infof(`Exporting stackdriver metrics with label "id"=%q`, id)

	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ReportingInterval:       1 * time.Minute,
		DefaultMonitoringLabels: m.labels,
		MonitoringClientOptions: []option.ClientOption{option.WithTokenSource(ts)},
	})
	if err != nil {
		return fmt.Errorf("failed to create stackdriver exporter: %v", err)
	}
	m.exporter = exporter
	return m.exporter.StartMetricsExporter()
}

// EndExport must be called to ensure all metrics are exported
func (m *ExportingClient) EndExport(ctx context.Context) {
	m.exporter.StopMetricsExporter()
	m.exporter.Flush()
}

// metric is a numeric value that is exported to the monitoring service.
type metric struct {
	startTime *timestamppb.Timestamp
	value     int64
}
