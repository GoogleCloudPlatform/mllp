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

package monitoring

import (
	"testing"

	"go.opencensus.io/metric/metricexport"
	"go.opencensus.io/metric/test"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"contrib.go.opencensus.io/exporter/stackdriver"
)

func TestClient(t *testing.T) {
	cl := &ExportingClient{
		projectID: "my-project-id",
		labels:    &stackdriver.Labels{},
		counters:  make(map[string]*stats.Int64Measure),
		latencies: make(map[string]*stats.Float64Measure),
	}
	cl.labels.Set("job", "mllp_adapter", "")
	cl.labels.Set("instance", "instance1", "")
	cl.labels.Set("zone", "zone1", "")

	metricreader := metricexport.NewReader()
	exporter := test.NewExporter(metricreader)

	cl.NewCounter("test-counter", "")
	cl.IncCounter("test-counter")
	cl.IncCounter("test-counter")
	cl.IncCounter("test-counter")

	cl.NewLatency("test-latency", "")
	cl.AddLatency("test-latency", 20)
	cl.AddLatency("test-latency", 100)
	cl.AddLatency("test-latency", 130)
	exporter.ReadAndExport()

	rows, err := view.RetrieveData(metricPrefix + "test-counter")
	if err != nil || len(rows) == 0 {
		t.Fatalf("Failed to get counter")
	}

	c, ok := rows[0].Data.(*view.CountData)
	if !ok {
		t.Errorf("want CountData, got %+v", rows[0].Data)
	}
	if c.Value != 3 {
		t.Errorf("Wrong counter result, expected 3, got %v", c.Value)
	}

	rows, err = view.RetrieveData(metricPrefix + "test-latency")
	if err != nil || len(rows) == 0 {
		t.Fatalf("Failed to get latency")
	}
	d, ok := rows[0].Data.(*view.DistributionData)
	if !ok {
		t.Errorf("want DistributionData, got %+v", rows[0].Data)
	}

	wantDistribution := &view.DistributionData{
		Count:           3,
		Mean:            83.33333333333333,
		SumOfSquaredDev: 6466.666666666667,
	}
	//gotDistribution := d.Value.(*metricdata.Distribution)
	if d.Count != wantDistribution.Count {
		t.Errorf("Unexpected distribution, expecting Count = %v, got Count = %v", wantDistribution.Count, d.Count)
	}
	if d.Mean != wantDistribution.Mean {
		t.Errorf("Unexpected distribution, expecting Mean = %v, got Mean = %v", wantDistribution.Mean, d.Mean)
	}
	if d.SumOfSquaredDev != wantDistribution.SumOfSquaredDev {
		t.Errorf("Unexpected distribution, expecting SumOfSquaredDev = %v, got SumOfSquaredDev = %v", wantDistribution.SumOfSquaredDev, d.SumOfSquaredDev)
	}
}
