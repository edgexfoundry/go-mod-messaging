//
// Copyright (c) 2019 Intel Corporation
// Copyright (c) 2022-2023 IOTech Ltd
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
	"encoding/base64"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRequestId     = "3ab0e022-464b-4bfe-bf7f-b0154093ddad"
	testCorrelationId = "fa1def22-96de-4d44-8811-00333438c8e3"
	testPayload       = `{"data" : "myData"}`
)

func TestNewMessageEnvelope(t *testing.T) {
	// lint:ignore SA1029 legacy
	// nolint:staticcheck // See golangci-lint #741
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, testCorrelationId)
	// lint:ignore SA1029 legacy
	// nolint:staticcheck // See golangci-lint #741
	ctx = context.WithValue(ctx, common.ContentType, common.ContentTypeJSON)

	envelope := NewMessageEnvelope(testPayload, ctx)

	assert.Equal(t, common.ApiVersion, envelope.ApiVersion)
	assert.Equal(t, testCorrelationId, envelope.CorrelationID)
	assert.Equal(t, common.ContentTypeJSON, envelope.ContentType)
	assert.Equal(t, testPayload, envelope.Payload)
	assert.Empty(t, envelope.QueryParams)
}

func TestNewMessageEnvelopeWithEnv(t *testing.T) {
	// lint:ignore SA1029 legacy
	// nolint:staticcheck // See golangci-lint #741
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, testCorrelationId)
	// lint:ignore SA1029 legacy
	// nolint:staticcheck // See golangci-lint #741
	ctx = context.WithValue(ctx, common.ContentType, common.ContentTypeJSON)

	_ = os.Setenv(envMsgBase64Payload, common.ValueTrue)
	defer os.Setenv(envMsgBase64Payload, common.ValueFalse)
	envelope := NewMessageEnvelope(testPayload, ctx)

	assert.Equal(t, common.ApiVersion, envelope.ApiVersion)
	assert.Equal(t, testCorrelationId, envelope.CorrelationID)
	assert.Equal(t, common.ContentTypeJSON, envelope.ContentType)
	assert.Equal(t, []byte(testPayload), envelope.Payload)
	assert.Empty(t, envelope.QueryParams)
}

func TestNewMessageEnvelopeEmpty(t *testing.T) {
	envelope := NewMessageEnvelope(nil, context.Background())

	assert.Equal(t, common.ApiVersion, envelope.ApiVersion)
	assert.Empty(t, envelope.RequestID)
	assert.Empty(t, envelope.CorrelationID)
	assert.Empty(t, envelope.ContentType)
	assert.Empty(t, envelope.Payload)
	assert.Zero(t, envelope.ErrorCode)
	assert.Empty(t, envelope.QueryParams)
}

func TestNewMessageEnvelopeForRequest(t *testing.T) {
	expectedQueryParams := map[string]string{"foo": "bar"}
	emptyQueryParams := map[string]string{}

	tests := []struct {
		name          string
		queryParams   map[string]string
		expectedEmpty bool
	}{
		{"valid - normal queryParams map", expectedQueryParams, false},
		{"valid - empty queryParams map", emptyQueryParams, true},
		{"valid - nil queryParams", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envelope := NewMessageEnvelopeForRequest(testPayload, tt.queryParams)

			assert.NotEmpty(t, envelope.RequestID)
			assert.NotEmpty(t, envelope.CorrelationID)
			assert.Equal(t, common.ApiVersion, envelope.ApiVersion)
			assert.Equal(t, testPayload, envelope.Payload)
			assert.Equal(t, common.ContentTypeJSON, envelope.ContentType)
			assert.Equal(t, 0, envelope.ErrorCode)
			assert.NotNil(t, envelope.QueryParams)
			if tt.expectedEmpty {
				assert.Empty(t, envelope.QueryParams)
				return
			}

			assert.Equal(t, expectedQueryParams, envelope.QueryParams)
		})
	}
}

func TestNewMessageEnvelopeForResponse(t *testing.T) {
	invalidUUID := "123456"

	tests := []struct {
		name          string
		correlationId string
		requestId     string
		contentType   string
		expectedError bool
	}{
		{"valid", testCorrelationId, testRequestId, common.ContentTypeJSON, false},
		{"invalid - CorrelationID is not in UUID format", invalidUUID, testRequestId, common.ContentTypeJSON, true},
		{"invalid - invalid requestID is not in UUID format", testCorrelationId, invalidUUID, common.ContentTypeJSON, true},
		{"invalid - ContentType is empty", testCorrelationId, testRequestId, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envelope, err := NewMessageEnvelopeForResponse(testPayload, tt.requestId, tt.correlationId, tt.contentType)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, testRequestId, envelope.RequestID)
			assert.Equal(t, testCorrelationId, envelope.CorrelationID)
			assert.Equal(t, common.ApiVersion, envelope.ApiVersion)
			assert.Equal(t, testPayload, envelope.Payload)
			assert.Equal(t, common.ContentTypeJSON, envelope.ContentType)
			assert.Equal(t, 0, envelope.ErrorCode)
			assert.NotNil(t, envelope.QueryParams)
		})
	}
}

func TestNewMessageEnvelopeFromJSON(t *testing.T) {
	invalidUUID := "123456"
	validEnvelope := testMessageEnvelope()
	validNoCorrelationIDEnvelope := validEnvelope
	validNoCorrelationIDEnvelope.CorrelationID = ""
	invalidApiVersionEnvelope := validEnvelope
	invalidApiVersionEnvelope.ApiVersion = "v1"
	invalidRequestIDEnvelope := validEnvelope
	invalidRequestIDEnvelope.RequestID = invalidUUID
	invalidCorrelationIDEnvelope := validEnvelope
	invalidCorrelationIDEnvelope.CorrelationID = invalidUUID
	invalidContentTypeEnvelope := validEnvelope
	invalidContentTypeEnvelope.ContentType = ""

	tests := []struct {
		name          string
		envelope      MessageEnvelope
		expectedError bool
	}{
		{"valid", validEnvelope, false},
		{"valid - CorrelationID is not set", validNoCorrelationIDEnvelope, false},
		{"invalid - API version not 'v2'", invalidApiVersionEnvelope, true},
		{"invalid - RequestID is not UUID format", invalidRequestIDEnvelope, true},
		{"invalid - CorrelationID is not UUID format", invalidCorrelationIDEnvelope, true},
		{"invalid - ContentType is not application/json", invalidContentTypeEnvelope, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := json.Marshal(tt.envelope)
			require.NoError(t, err)

			envelope, err := NewMessageEnvelopeFromJSON(payload)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, testRequestId, envelope.RequestID)
			assert.NotEmpty(t, testCorrelationId, envelope.CorrelationID)
			assert.Equal(t, common.ApiVersion, envelope.ApiVersion)
			assert.Equal(t, testPayload, envelope.Payload)
			assert.Equal(t, common.ContentTypeJSON, envelope.ContentType)
			assert.Equal(t, 0, envelope.ErrorCode)
			assert.NotNil(t, envelope.QueryParams)
		})
	}
}

func TestNewMessageEnvelopeWithError(t *testing.T) {
	expectedPayload := `error: something failed`

	envelope := NewMessageEnvelopeWithError(testRequestId, expectedPayload)

	assert.NotEmpty(t, testCorrelationId, envelope.CorrelationID)
	assert.Equal(t, common.ApiVersion, envelope.ApiVersion)
	assert.Equal(t, testRequestId, envelope.RequestID)
	assert.Equal(t, 1, envelope.ErrorCode)
	assert.Equal(t, expectedPayload, envelope.Payload)
	assert.Equal(t, common.ContentTypeText, envelope.ContentType)
	assert.Empty(t, envelope.QueryParams)
}

func TestMessageEnvelopeJSON(t *testing.T) {
	expected := testMessageEnvelope()
	expected.QueryParams = make(map[string]string)
	expected.QueryParams["q1"] = "v1"

	data, err := json.Marshal(expected)
	require.NoError(t, err)

	actual, err := NewMessageEnvelopeFromJSON(data)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func testMessageEnvelope() MessageEnvelope {
	return MessageEnvelope{
		CorrelationID: testCorrelationId,
		Versionable:   commonDTO.NewVersionable(),
		RequestID:     testRequestId,
		ErrorCode:     0,
		Payload:       testPayload,
		ContentType:   common.ContentTypeJSON,
	}
}

func TestGetMsgPayloadFromBytesToStruct(t *testing.T) {
	testStructPayload := responses.EventResponse{}
	testStructBytesPayload, err := json.Marshal(testStructPayload)
	require.NoError(t, err)
	testBase64Payload := base64.StdEncoding.EncodeToString(testStructBytesPayload)

	msgEnvelope := testMessageEnvelope()
	msgEnvelope.Payload = testBase64Payload

	res, err := GetMsgPayload[responses.EventResponse](msgEnvelope)
	assert.NoError(t, err)
	assert.IsType(t, "responses.EventResponse", reflect.TypeOf(res).String())
	assert.Equal(t, testStructPayload, res)
}

func TestGetMsgPayloadFromBytesToBytes(t *testing.T) {
	testBytesPayload := []byte(testPayload)
	testBase64Payload := base64.StdEncoding.EncodeToString(testBytesPayload)

	msgEnvelope := testMessageEnvelope()
	msgEnvelope.Payload = testBase64Payload

	res, err := GetMsgPayload[[]byte](msgEnvelope)
	assert.NoError(t, err)
	assert.IsType(t, "[]uint8", reflect.TypeOf(res).String())
	assert.Equal(t, testBytesPayload, res)
}
