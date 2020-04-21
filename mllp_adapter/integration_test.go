// Copyright 2019 Google LLC
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

package integration_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"flag"
	log "github.com/golang/glog"
	"github.com/GoogleCloudPlatform/mllp/mllp_adapter/mllp"

	healthcare "google.golang.org/api/healthcare/v1"
)

var (
	hl7V2ProjectID      = flag.String("hl7_v2_project_id", "", "Project ID that owns the healthcare dataset")
	hl7V2LocationID     = flag.String("hl7_v2_location_id", "us-central1", "ID of Cloud Location where the healthcare dataset is stored")
	hl7V2DatasetID      = flag.String("hl7_v2_dataset_id", "", "ID of the healthcare dataset")
	hl7V2StoreID        = flag.String("hl7_v2_store_id", "", "ID of the HL7V2 store inside the healthcare dataset")
	pubsubTopic         = flag.String("pubsub_topic", "", "Pubsub topic associated with the HL7V2 store")
	mllpAdapterAddr     = flag.String("mllp_adapter_addr", "", "Address of the MLLP adapter")
	fakeHospitalStoreID = flag.String("fake_hospital_store_id", "", "ID of the HL7V2 store which stores the outgoing message from the MLLP adapter")

	datasetService    *healthcare.ProjectsLocationsDatasetsService
	hl7V2StoreService *healthcare.ProjectsLocationsDatasetsHl7V2StoresService
)

func TestSendToMLLPAdapter(t *testing.T) {
	ctx := context.Background()
	message := "MSH|^~\\&|FROM_APP|FROM_FACILITY_A|TO_APP|TO_FACILITY|20150503223000||ADT^A01|20150503223000|T|2.3|||AL||44|ASCII\rPID||843125^^^^MRN|21004053^^^^MRN~2269030303^^^^NHSNMBR"
	ack, err := sendHL7V2MessageToMLLPAdapter(t, message)
	if err != nil {
		t.Fatalf("Send message to the MLLP adapter: %v", err)
	}
	if !bytes.Contains(ack, []byte("MSA|AA")) {
		t.Errorf("Failed to find \"AA\"(Application Accept) in the returned message: %s", ack)
	}

	storeEncodedMessage, err := getMessageInHL7V2Store(ctx, *hl7V2StoreID, "sendFacility = \"FROM_FACILITY_A\"", 1)
	if err != nil {
		t.Fatalf("Get message in the HL7V2 store: %v", err)
	}
	storeMessage, err := base64.StdEncoding.DecodeString(storeEncodedMessage)
	if err != nil {
		t.Fatalf("Failed to decode message returned from the HL7V2 store: %v", err)
	}
	if string(storeMessage) != message {
		t.Fatalf("Message in the HL7V2 store %q is different than the message sent to the MLLP adapter %q", storeMessage, message)
	}
}

func TestSendToMLLPAdapterInvalidMessage(t *testing.T) {
	ctx := context.Background()
	message := "MSH|^~\\&|FROM_APP|FROM_FACILITY_B|TO_APP|TO_FACILITY|20150503223000||ADT^A01|20150503223000|T|2.3|||AL||44|ASCII\nPID||843125^^^^MRN|21004053^^^^MRN~2269030303^^^^NHSNMBR"
	nack, err := sendHL7V2MessageToMLLPAdapter(t, message)
	if err != nil {
		t.Fatalf("Send message to the MLLP adapter: %v", err)
	}
	if !bytes.Contains(nack, []byte("MSA|AE")) {
		t.Errorf("Failed to find \"AE\"(Application Error) in the returned message: %s", nack)
	}

	if _, err := getMessageInHL7V2Store(ctx, *hl7V2StoreID, "sendFacility = \"FROM_FACILITY_B\"", 0); err != nil {
		t.Fatalf("Get message in the HL7V2 store: %v", err)
	}
}

func TestMLLPAdapterListenToPubSub(t *testing.T) {
	ctx := context.Background()
	// Create a new meesage in the HL7V2 store, the mllp adapter listening on
	// the store's pubsub topic will forward the message to the fake hospital.
	message := []byte("MSH|^~\\&|FROM_APP|FROM_FACILITY_C|TO_APP|TO_FACILITY|20150503223000||ADT^A01|20150503223000|T|2.3|||AL||44|ASCII\rPID||843125^^^^MRN|21004053^^^^MRN~2269030303^^^^NHSNMBR")
	_, err := createMessageInHL7V2Store(context.Background(), *hl7V2StoreID, base64.StdEncoding.EncodeToString(message))
	if err != nil {
		t.Error(err)
	}
	storeEncodedMessage, err := getMessageInHL7V2Store(ctx, *fakeHospitalStoreID, "sendFacility = \"FROM_FACILITY_C\"", 1)
	if err != nil {
		t.Fatalf("Get message in the fake hospital store: %v", err)
	}
	storeMessage, err := base64.StdEncoding.DecodeString(storeEncodedMessage)
	if err != nil {
		t.Fatalf("Failed to decode message returned from the fake hospital store: %v", err)
	}
	if string(storeMessage) != string(message) {
		t.Fatalf("Message sent by the MLLP adapter %q is different than the messaged created in the HL7V2 store %q", storeMessage, message)
	}

}

func getMessageInHL7V2Store(ctx context.Context, storeID, filter string, wantCount int) (string, error) {
	for i := 0; i < 6; i++ {
		time.Sleep(10 * time.Second)
		lr, err := hl7V2StoreService.Messages.List(hl7V2StoreName(storeID)).Filter(filter).View("RAW_ONLY").Context(ctx).Do()
		if err != nil {
			return "", fmt.Errorf("failed to list messages with filter %q: %v", filter, err)
		}
		if got := len(lr.Hl7V2Messages); got != wantCount {
			if i < 6 {
				continue
			}
			return "", fmt.Errorf("listing messages with filter %q returned %d messages, want %d", filter, got, wantCount)
		}
		if wantCount == 0 {
			return "", nil
		}
		return lr.Hl7V2Messages[0].Data, nil
	}
	return "", nil
}

func sendHL7V2MessageToMLLPAdapter(t *testing.T, hl7V2Message string) ([]byte, error) {
	conn, err := net.Dial("tcp", *mllpAdapterAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the MLLP adapter at %s: %v", *mllpAdapterAddr, err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Errorf("failed to clean up connection: %v", err)
		}
	}()

	if err := mllp.WriteMsg(conn, []byte(hl7V2Message)); err != nil {
		return nil, fmt.Errorf("failed to write message to the MLLP adapter: %v", err)
	}
	ack, err := mllp.ReadMsg(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read ACK/NACK message from the MLLP adapter: %v", err)
	}
	return ack, nil
}

func createMessageInHL7V2Store(ctx context.Context, storeID, hl7V2Message string) (*healthcare.Message, error) {
	req := &healthcare.CreateMessageRequest{
		Message: &healthcare.Message{
			Data: string(hl7V2Message),
		},
	}
	resp, err := hl7V2StoreService.Messages.Create(hl7V2StoreName(storeID), req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create HL7V2 message(%s) in the HL7V2 store: %v", hl7V2Message, err)
	}
	return resp, nil
}

func createDataset(ctx context.Context) error {
	resp, err := datasetService.Create(projectName(), &healthcare.Dataset{}).DatasetId(*hl7V2DatasetID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create dataset: %v", err)
	}
	if err := waitForOperation(resp.Name); err != nil {
		return err
	}
	return nil
}

func waitForOperation(name string) error {
	ctx := context.Background()
	for {
		op, err := datasetService.Operations.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("failed to wait for operation %s: %v", name, err)
		}
		if op.Done {
			return nil
		}
	}
}

func deleteDataset(ctx context.Context) error {
	_, err := datasetService.Delete(datasetName()).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete dataset: %v", err)
	}
	return nil
}

func createHL7V2Store(ctx context.Context, storeID, pubsubTopic string) error {
	store := &healthcare.Hl7V2Store{}
	if pubsubTopic != "" {
		store.NotificationConfigs = []*healthcare.Hl7V2NotificationConfig{
			&healthcare.Hl7V2NotificationConfig{
				PubsubTopic: fmt.Sprintf("projects/%s/topics/%s", *hl7V2ProjectID, pubsubTopic),
			},
		}
	}
	_, err := hl7V2StoreService.Create(datasetName(), store).Hl7V2StoreId(storeID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create the HL7V2 store %q: %v", storeID, err)
	}
	return nil
}

func deleteHL7V2Store(ctx context.Context, storeID string) error {
	_, err := hl7V2StoreService.Delete(hl7V2StoreName(storeID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete the HL7V2 store %q: %v", storeID, err)
	}
	return nil
}

func projectName() string {
	return fmt.Sprintf("projects/%s/locations/%s", *hl7V2ProjectID, *hl7V2LocationID)
}

func datasetName() string {
	return fmt.Sprintf("%s/datasets/%s", projectName(), *hl7V2DatasetID)
}

func hl7V2StoreName(storeID string) string {
	return fmt.Sprintf("%s/hl7V2Stores/%s", datasetName(), storeID)
}

func runTests(m *testing.M) (int, error) {
	ctx := context.Background()
	healthcareService, err := healthcare.NewService(ctx)
	if err != nil {
		return 0, fmt.Errorf("healthcare.NewService(ctx) returned unexpected error: %v", err)
	}
	datasetService = healthcareService.Projects.Locations.Datasets
	hl7V2StoreService = datasetService.Hl7V2Stores
	if err := createDataset(ctx); err != nil {
		return 0, err
	}
	if err := createHL7V2Store(ctx, *hl7V2StoreID, *pubsubTopic); err != nil {
		return 0, err
	}
	if err := createHL7V2Store(ctx, *fakeHospitalStoreID, ""); err != nil {
		return 0, err
	}

	// Execute the tests.
	res := m.Run()

	if err := deleteHL7V2Store(ctx, *hl7V2StoreID); err != nil {
		return 0, err
	}
	if err := deleteHL7V2Store(ctx, *fakeHospitalStoreID); err != nil {
		return 0, err
	}
	if err := deleteDataset(ctx); err != nil {
		return 0, err
	}

	return res, nil
}

func TestMain(m *testing.M) {
	flag.Parse()
	res, err := runTests(m)
	if err != nil {
		log.Exitf("Error running tests: %v", err)
	}
	os.Exit(res)
}
