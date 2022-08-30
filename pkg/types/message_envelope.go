//
// Copyright (c) 2019 Intel Corporation
// Copyright (c) 2022 IOTech Ltd
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
//

package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const (
	ApiVersion      = "v2"
	CorrelationID   = "X-Correlation-ID"
	ContentType     = "Content-Type"
	ContentTypeJSON = "application/json"
	ContentTypeText = "text/plain"
)

// MessageEnvelope is the data structure for messages. It wraps the generic message payload with attributes.
type MessageEnvelope struct {
	// ReceivedTopic is the topic that the message was received on.
	ReceivedTopic string
	// CorrelationID is an object id to identify the envelope.
	CorrelationID string
	// ApiVersion shows the API version in message envelope.
	ApiVersion string
	// RequestID is an object id to identify the request.
	RequestID string
	// ErrorCode provides the indication of error. '0' indicates no error, '1' indicates error.
	// Additional codes may be added in the future. If non-0, the payload will contain the error.
	ErrorCode int
	// Payload is byte representation of the data being transferred.
	Payload []byte
	// ContentType is the marshaled type of payload, i.e. application/json, application/xml, application/cbor, etc
	ContentType string
	// QueryParams is optionally provided key/value pairs.
	QueryParams map[string]string
}

// NewMessageEnvelope creates a new MessageEnvelope for the specified payload with attributes from the specified context
func NewMessageEnvelope(payload []byte, ctx context.Context) MessageEnvelope {
	envelope := MessageEnvelope{
		ApiVersion:    ApiVersion,
		CorrelationID: fromContext(ctx, CorrelationID),
		ContentType:   fromContext(ctx, ContentType),
		Payload:       payload,
		QueryParams:   make(map[string]string),
	}

	return envelope
}

// NewMessageEnvelopeForRequest creates a new MessageEnvelope for sending request to EdgeX via internal
// MessageBus to target Device Service. Used when request is from internal App Service via command client.
func NewMessageEnvelopeForRequest(payload []byte, queryParams map[string]string) MessageEnvelope {
	envelope := MessageEnvelope{
		CorrelationID: uuid.NewString(),
		ApiVersion:    ApiVersion,
		RequestID:     uuid.NewString(),
		ErrorCode:     0,
		Payload:       payload,
		ContentType:   ContentTypeJSON,
		QueryParams:   make(map[string]string),
	}

	if len(queryParams) > 0 {
		envelope.QueryParams = queryParams
	}

	return envelope
}

// NewMessageEnvelopeForResponse creates a new MessageEnvelope for sending response from Device Service back to Core Command.
func NewMessageEnvelopeForResponse(payload []byte, requestId string, correlationId string, contentType string) (MessageEnvelope, error) {
	if _, err := uuid.Parse(requestId); err != nil {
		return MessageEnvelope{}, err
	}
	if _, err := uuid.Parse(correlationId); err != nil {
		return MessageEnvelope{}, err
	}
	if contentType == "" {
		return MessageEnvelope{}, errors.New("ContentType is empty")
	}

	envelope := MessageEnvelope{
		CorrelationID: correlationId,
		ApiVersion:    ApiVersion,
		RequestID:     requestId,
		ErrorCode:     0,
		Payload:       payload,
		ContentType:   contentType,
		QueryParams:   make(map[string]string),
	}

	return envelope, nil
}

// NewMessageEnvelopeFromJSON creates a new MessageEnvelope by decoding the message payload
// received from external MQTT in order to send request via internal MessageBus.
func NewMessageEnvelopeFromJSON(message []byte) (MessageEnvelope, error) {
	var envelope MessageEnvelope
	err := json.Unmarshal(message, &envelope)
	if err != nil {
		return MessageEnvelope{}, err
	}

	if envelope.ApiVersion != ApiVersion {
		return MessageEnvelope{}, errors.New("api version 'v2' is required")
	}

	if _, err = uuid.Parse(envelope.RequestID); err != nil {
		return MessageEnvelope{}, fmt.Errorf("error parsing RequestID: %s", err.Error())
	}

	if _, err = uuid.Parse(envelope.CorrelationID); err != nil {
		if envelope.CorrelationID != "" {
			return MessageEnvelope{}, fmt.Errorf("error parsing CorrelationID: %s", err.Error())
		}

		envelope.CorrelationID = uuid.NewString()
	}

	if envelope.ContentType != ContentTypeJSON {
		return envelope, errors.New("ContentType is not application/json")
	}

	if envelope.QueryParams == nil {
		envelope.QueryParams = make(map[string]string)
	}

	return envelope, nil
}

// NewMessageEnvelopeWithError creates a new MessageEnvelope with ErrorCode set to 1 indicating there's error
// and the payload contains message string about the error.
func NewMessageEnvelopeWithError(requestId string, errorMessage string) MessageEnvelope {
	return MessageEnvelope{
		CorrelationID: uuid.NewString(),
		ApiVersion:    ApiVersion,
		RequestID:     requestId,
		ErrorCode:     1,
		Payload:       []byte(errorMessage),
		ContentType:   ContentTypeText,
		QueryParams:   make(map[string]string),
	}
}

func fromContext(ctx context.Context, key string) string {
	hdr, ok := ctx.Value(key).(string)
	if !ok {
		hdr = ""
	}
	return hdr
}
