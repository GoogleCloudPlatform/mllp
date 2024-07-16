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

// Package mllp contains functionality for encoding and decoding HL7 messages for transmission using the MLLP protocol.
// See here for the specification:
// http://www.hl7.org/documentcenter/public_temp_670395EE-1C23-BA17-0CD218684D5B3C71/wg/inm/mllp_transport_specification.PDF
package mllp

import (
	"bufio"
	"fmt"
	"io"

	log "github.com/golang/glog"
)

const (
	startBlock = '\x0b'
	endBlock   = '\x1c'
	cr         = '\x0d'
)

// WriteMsg wraps an HL7 message in the start block, end block, and carriage return bytes
// required for MLLP transmission and then writes the wrapped message to writer.
func WriteMsg(writer io.Writer, msg []byte) error {
	if _, err := writer.Write([]byte{startBlock}); err != nil {
		return fmt.Errorf("writing message: %v", err)
	}
	if _, err := writer.Write(msg); err != nil {
		return fmt.Errorf("writing message: %v", err)
	}
	if _, err := writer.Write([]byte{endBlock, cr}); err != nil {
		return fmt.Errorf("writing message: %v", err)
	}
	return nil
}

// MessageReader consumes MLLP messages from a stream.
type MessageReader struct {
	r *bufio.Reader
}

// NewMessageReader to unwrap MLLP messages the provided stream.
func NewMessageReader(r io.Reader) *MessageReader {
	return &MessageReader{r: bufio.NewReader(r)}
}

// Next message in the reader. Unwraps the inner message by removing the start
// block, end block, and carriage return bytes.
func (mr *MessageReader) Next() ([]byte, error) {
	data, err := mr.r.ReadBytes(startBlock)
	if err != nil {
		return nil, err
	}
	if len(data) > 1 {
		log.Infof("dropped %d bytes before start of message", len(data)-1)
	}
	// Read everything up to the endBlock byte.
	rawMsg, err := mr.r.ReadBytes(endBlock)
	if err != nil {
		return nil, err
	}
	// Read one more byte for the carriage return.
	lastByte, err := mr.r.ReadByte()
	if err != nil {
		return nil, err
	}
	if lastByte != cr {
		if err := mr.r.UnreadByte(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("message ends with %c, want %c", lastByte, cr)
	}
	return rawMsg[:len(rawMsg)-1], nil
}

// ReadMsg from reader and removes the start block, end block, and carriage return bytes.
// The reader must return a single message, any trailing bytes may be consumed.
func ReadMsg(r io.Reader) ([]byte, error) {
	return NewMessageReader(r).Next()
}
