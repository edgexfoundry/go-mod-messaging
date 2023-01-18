//
// Copyright (C) 2022 IOTech Ltd
// Copyright (c) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-messaging/v3/messaging/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
)

const (
	testDeviceName                = "test-device"
	testCommandName               = "test-command"
	testQueryRequestTopic         = "test/commandquery/request"
	testCommandRequestTopicPrefix = "test/command/request"
)

var expectedRequestID = uuid.NewString()
var expectedCorrelationID = uuid.NewString()
var errorResponse = types.NewMessageEnvelopeWithError(expectedRequestID, "request timed out")

func TestCommandClient_AllDeviceCoreCommands(t *testing.T) {
	responseDTO := responses.NewMultiDeviceCoreCommandsResponse(expectedRequestID, "", http.StatusOK, 0, nil)
	responseBytes, err := json.Marshal(responseDTO)
	require.NoError(t, err)

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, expectedRequestID, expectedCorrelationID, common.ContentTypeJSON)
	require.NoError(t, err)

	tests := []struct {
		Name                 string
		ExpectedResponse     *types.MessageEnvelope
		ExpectedRequestError error
		ExpectError          bool
	}{
		{"valid", &responseEnvelope, nil, false},
		{"request error", nil, errors.New("timed out"), true},
		{"response error", &errorResponse, nil, true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := getCommandClientWithMockMessaging(t, test.ExpectedResponse, test.ExpectedRequestError)

			res, err := client.AllDeviceCoreCommands(context.Background(), 0, 20)

			if test.ExpectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.IsType(t, res, responses.MultiDeviceCoreCommandsResponse{})
			assert.Equal(t, res.RequestId, expectedRequestID)
		})
	}
}

func TestCommandClient_DeviceCoreCommandsByDeviceName(t *testing.T) {
	responseDTO := responses.NewDeviceCoreCommandResponse(expectedRequestID, "", http.StatusOK, dtos.DeviceCoreCommand{})
	responseBytes, err := json.Marshal(responseDTO)
	require.NoError(t, err)

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, expectedRequestID, expectedCorrelationID, common.ContentTypeJSON)
	require.NoError(t, err)

	tests := []struct {
		Name                 string
		ExpectedResponse     *types.MessageEnvelope
		ExpectedRequestError error
		ExpectError          bool
	}{
		{"valid", &responseEnvelope, nil, false},
		{"request error", nil, errors.New("timed out"), true},
		{"response error", &errorResponse, nil, true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := getCommandClientWithMockMessaging(t, test.ExpectedResponse, test.ExpectedRequestError)

			res, err := client.DeviceCoreCommandsByDeviceName(context.Background(), testDeviceName)

			if test.ExpectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.IsType(t, res, responses.DeviceCoreCommandResponse{})
			assert.Equal(t, res.RequestId, expectedRequestID)
		})
	}
}

func TestCommandClient_IssueGetCommandByName(t *testing.T) {
	responseDTO := responses.NewEventResponse(expectedRequestID, "", http.StatusOK, dtos.Event{})
	responseBytes, err := json.Marshal(responseDTO)
	require.NoError(t, err)

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, expectedRequestID, expectedCorrelationID, common.ContentTypeJSON)
	require.NoError(t, err)

	tests := []struct {
		Name                 string
		ExpectedResponse     *types.MessageEnvelope
		ExpectedRequestError error
		ExpectError          bool
	}{
		{"valid", &responseEnvelope, nil, false},
		{"request error", nil, errors.New("timed out"), true},
		{"response error", &errorResponse, nil, true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := getCommandClientWithMockMessaging(t, test.ExpectedResponse, test.ExpectedRequestError)

			res, err := client.IssueGetCommandByName(context.Background(), testDeviceName, testCommandName, false, true)

			if test.ExpectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.IsType(t, res, &responses.EventResponse{})
			assert.Equal(t, res.RequestId, expectedRequestID)
		})
	}
}

func TestCommandClient_IssueGetCommandByNameWithQueryParams(t *testing.T) {
	responseDTO := responses.NewEventResponse(expectedRequestID, "", http.StatusOK, dtos.Event{})
	responseBytes, err := json.Marshal(responseDTO)
	require.NoError(t, err)

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, expectedRequestID, expectedCorrelationID, common.ContentTypeJSON)
	require.NoError(t, err)

	tests := []struct {
		Name                 string
		ExpectedResponse     *types.MessageEnvelope
		ExpectedRequestError error
		ExpectError          bool
	}{
		{"valid", &responseEnvelope, nil, false},
		{"request error", nil, errors.New("timed out"), true},
		{"response error", &errorResponse, nil, true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := getCommandClientWithMockMessaging(t, test.ExpectedResponse, test.ExpectedRequestError)

			res, err := client.IssueGetCommandByNameWithQueryParams(context.Background(), testDeviceName, testCommandName, nil)

			if test.ExpectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.IsType(t, res, &responses.EventResponse{})
			assert.Equal(t, res.RequestId, expectedRequestID)
		})
	}
}

func TestCommandClient_IssueSetCommandByName(t *testing.T) {
	responseDTO := commonDTO.NewBaseResponse(expectedRequestID, "", http.StatusOK)
	responseBytes, err := json.Marshal(responseDTO)
	require.NoError(t, err)

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, expectedRequestID, expectedCorrelationID, common.ContentTypeJSON)
	require.NoError(t, err)

	tests := []struct {
		Name                 string
		ExpectedResponse     *types.MessageEnvelope
		ExpectedRequestError error
		ExpectError          bool
	}{
		{"valid", &responseEnvelope, nil, false},
		{"request error", nil, errors.New("timed out"), true},
		{"response error", &errorResponse, nil, true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := getCommandClientWithMockMessaging(t, test.ExpectedResponse, test.ExpectedRequestError)

			res, err := client.IssueSetCommandByName(context.Background(), testDeviceName, testCommandName, nil)

			if test.ExpectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.IsType(t, res, commonDTO.BaseResponse{})
			assert.Equal(t, res.RequestId, expectedRequestID)
		})
	}
}

func TestCommandClient_IssueSetCommandByNameWithObject(t *testing.T) {
	responseDTO := commonDTO.NewBaseResponse(expectedRequestID, "", http.StatusOK)
	responseBytes, err := json.Marshal(responseDTO)
	require.NoError(t, err)

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, expectedRequestID, expectedCorrelationID, common.ContentTypeJSON)
	require.NoError(t, err)

	tests := []struct {
		Name                 string
		ExpectedResponse     *types.MessageEnvelope
		ExpectedRequestError error
		ExpectError          bool
	}{
		{"valid", &responseEnvelope, nil, false},
		{"request error", nil, errors.New("timed out"), true},
		{"response error", &errorResponse, nil, true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := getCommandClientWithMockMessaging(t, test.ExpectedResponse, test.ExpectedRequestError)

			res, err := client.IssueSetCommandByNameWithObject(context.Background(), testDeviceName, testCommandName, nil)

			if test.ExpectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.IsType(t, res, commonDTO.BaseResponse{})
			assert.Equal(t, res.RequestId, expectedRequestID)
		})
	}
}

func getCommandClientWithMockMessaging(t *testing.T, expectedResponse *types.MessageEnvelope, expectedRequestError error) interfaces.CommandClient {
	mockMessageClient := &mocks.MessageClient{}
	mockMessageClient.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, expectedRequestError)

	topics := map[string]string{
		QueryRequestTopicPrefix:   testQueryRequestTopic,
		CommandRequestTopicPrefix: testCommandRequestTopicPrefix,
	}

	client, err := NewCommandClient(mockMessageClient, topics, 10*time.Second)
	require.NoError(t, err)

	return client
}
