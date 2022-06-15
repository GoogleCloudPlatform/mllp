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

// The mllp_adapter binary is a server that accepts HL7 messages over MLLP and
// forwards them to the Cloud HL7 service API.
package main

import (
	"context"
	"fmt"
	"os"

	"flag"
	
	log "github.com/golang/glog"
	"github.com/GoogleCloudPlatform/mllp/mllp_adapter/handler"
	"github.com/GoogleCloudPlatform/mllp/mllp_adapter/mllpreceiver"
	"github.com/GoogleCloudPlatform/mllp/mllp_adapter/mllpsender"
	"github.com/GoogleCloudPlatform/mllp/shared/healthapiclient"
	"github.com/GoogleCloudPlatform/mllp/shared/monitoring"
	"github.com/GoogleCloudPlatform/mllp/shared/pubsub"
)

var (
	// 2575 is the default port for HL7 over TCP
	// https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?search=2575
	port                  = flag.Int("port", 2575, "Port on which to listen for incoming MLLP connections")
	apiAddrPrefix         = flag.String("api_addr_prefix", "", "[Deprecated] Prefix of the Cloud Healthcare API, including scheme and version")
	mllpAddr              = flag.String("mllp_addr", "", "Target address for outgoing MLLP connections")
	receiverIP            = flag.String("receiver_ip", "", "IP address for incoming MLLP connections")
	pubsubProjectID       = flag.String("pubsub_project_id", "", "Project ID that owns the pubsub topic")
	pubsubSubscription    = flag.String("pubsub_subscription", "", "Pubsub subscription to read for notifications of new messages")
	hl7V2ProjectID        = flag.String("hl7_v2_project_id", "", "Project ID that owns the healthcare dataset")
	hl7V2LocationID       = flag.String("hl7_v2_location_id", "", "ID of Cloud Location where the healthcare dataset is stored")
	hl7V2DatasetID        = flag.String("hl7_v2_dataset_id", "", "ID of the healthcare dataset")
	hl7V2StoreID          = flag.String("hl7_v2_store_id", "", "ID of the HL7v2 store inside the healthcare dataset")
	logNACKedMsg          = flag.Bool("log_nacked_msg", false, "[Optional] Whether to log the contents of messages that receive NACK from the API. These error logs will contain sensitive data.")
	exportStats           = flag.Bool("export_stats", true, "[Optional] Whether to export stackdriver stats")
	credentials           = flag.String("credentials", "", "[Optional] Path to the credentials file (in JSON format). The default service account will be used if not provided.")
	checkPublishAttribute = flag.Bool("legacy_publish_attribute", false,
		"[Optional] Whether to check for the publish attribute when reading pubsub subscriptions. This attribute appears only in the notifications from messages.create method, and will be removed in a future release.")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	var mon *monitoring.ExportingClient
	if *exportStats {
		mon = monitoring.NewExportingClient()
		if err := mon.StartExport(ctx, *credentials); err != nil {
			return fmt.Errorf("failed to configure monitoring: %v", err)
		}

		defer mon.EndExport(ctx)
	}

	if *apiAddrPrefix != "" {
		log.Warningf("Flag --api_addr_prefix deprecated, API calls will be made to healthcare.googleapis.com/v1.")
	}
	apiClient, err := healthapiclient.NewHL7V2Client(ctx, *credentials, mon, *hl7V2ProjectID, *hl7V2LocationID, *hl7V2DatasetID, *hl7V2StoreID, *logNACKedMsg)
	if err != nil {
		return fmt.Errorf("failed to connect to HL7v2 API: %v", err)
	}

	if *pubsubProjectID == "" || *pubsubSubscription == "" {
		log.Infof("Either --pubsub_project_id or --pubsub_subscription is not provided, notifications of the new messages are not read and no outgoing messages will be sent to the target MLLP address.")
	} else {
		sender := mllpsender.NewSender(*mllpAddr, mon)
		handler := handler.New(mon, apiClient, sender, *checkPublishAttribute)
		go func() {
			err := pubsub.Listen(ctx, *credentials, handler, *pubsubProjectID, *pubsubSubscription)
			log.Errorf("MLLP Adapter: failed to connect to PubSub channel: %v", err)
			os.Exit(1)
		}()
	}

	if *receiverIP == "" {
		return fmt.Errorf("required flag value --receiver_ip not provided")
	}

	receiver, err := mllpreceiver.NewReceiver(*receiverIP, *port, apiClient, mon)
	if err != nil {
		return fmt.Errorf("failed to create MLLP receiver: %v", err)
	}

	go func() {
		if err := receiver.Run(); err != nil {
			log.Errorf("MLLP Adapter: failed to start MLLP receiver: %v", err)
			os.Exit(1)
		}
	}()

	select {}
}
