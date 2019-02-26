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

package util

import (
	"context"
	"strings"
	"testing"
)

const validParentPath string = "projects/p1/location/l1/datasets/ds1"

func TestGenerateHL7V2StoreName(t *testing.T) {
	expected := "projects/p1/locations/l1/datasets/ds1/hl7V2Stores/hl71"
	if name := GenerateHL7V2StoreName("p1", "l1", "ds1", "hl71"); name != expected {
		t.Errorf("GenerateHL7V2StoreName(\"p1\", \"l1\", \"ds1\", \"hl71\") => %v, expected %v", name, expected)
	}
}

func TestParseHL7V2MessageName_Success(t *testing.T) {
	msgName := GenerateHL7V2MessageName("p1", "l1", "ds1", "hl71", "msg1")
	pRef, lID, dID, hl7V2StoreID, msgID, err := ParseHL7V2MessageName(msgName)
	if pRef != "p1" || lID != "l1" || dID != "ds1" || hl7V2StoreID != "hl71" || msgID != "msg1" || err != nil {
		t.Errorf("ParseHL7V2MessageName(%v) => (%v, %v, %v, %v, %v, %v) expected (p1, l1, ds1, hl71, msg1, nil)", msgName, pRef, lID, dID, hl7V2StoreID, msgID, err)
	}
}

func TestGenerateFHIRStoreName(t *testing.T) {
	expected := "projects/p1/locations/l1/datasets/ds1/fhirStores/f1"
	if name := GenerateFHIRStoreName("p1", "l1", "ds1", "f1"); name != expected {
		t.Errorf("GenerateFHIRStoreName(\"p1\", \"l1\", \"ds1\", \"f1\") => %v, expected %v", name, expected)
	}
}

func TestParseHL7V2MessageName_Errors(t *testing.T) {
	tests := []struct {
		name, msgName string
	}{
		{
			"invalid parent name",
			"blahblah",
		},
		{
			"invalid HL7v2 stores component",
			validParentPath + strings.Join([]string{"invalid", "hl71", messagesPathComponent, "msg1"}, "/"),
		},
		{
			"invalid messages component",
			validParentPath + strings.Join([]string{hl7V2StoresPathComponent, "hl71", "invalid", "msg1"}, "/"),
		},
		{
			"missing message ID and messages component",
			validParentPath + strings.Join([]string{hl7V2StoresPathComponent, "hl71"}, "/"),
		},
	}

	for _, test := range tests {
		if pRef, lID, dID, hl7V2StoreID, msgID, err := ParseHL7V2MessageName(test.msgName); err == nil {
			t.Errorf("%v: ParseHL7V2MessageName(%v) => (%v, %v, %v, %v, %v, nil) expected non nil error", test.name, test.msgName, pRef, lID, dID, hl7V2StoreID, msgID)
		}
	}
}

func TestTokenSource_NoSuchFile(t *testing.T) {
	p := "some_file.json"
	_, err := TokenSource(context.Background(), p)
	if err == nil {
		t.Errorf("TokenSource(ctx, %s) got no error, expected one ", p)
	}
}
