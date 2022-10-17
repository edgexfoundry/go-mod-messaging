//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

const (
	testDeviceName                = "test-device"
	testCommandName               = "test-command"
	testQueryRequestTopic         = "test/commandquery/request"
	testQueryResponseTopic        = "test/commandquery/response"
	testCommandRequestTopicPrefix = "test/command/request"
	testCommandResponseTopic      = "test/command/response/#"
)

var testRequestID string
var testCorrelationID string

type mockMessageClient struct {
	signal chan struct{}
}

func (m mockMessageClient) Connect() error {
	//TODO implement me
	panic("implement me")
}

func (m mockMessageClient) Publish(message types.MessageEnvelope, topic string) error {
	testRequestID = message.RequestID
	testCorrelationID = message.CorrelationID

	m.signal <- struct{}{}
	return nil
}

func (m mockMessageClient) Subscribe(topics []types.TopicChannel, messageErrors chan error) error {
	return nil
}

func (m mockMessageClient) Disconnect() error {
	//TODO implement me
	panic("implement me")
}

func TestCommandClient_AllDeviceCoreCommands(t *testing.T) {
	mockMessageBus := mockMessageClient{signal: make(chan struct{})}
	topics := map[string]string{
		QueryRequestTopic:  testQueryRequestTopic,
		QueryResponseTopic: testQueryResponseTopic,
	}

	client, err := NewCommandClient(mockMessageBus, topics, 10*time.Second)
	require.NoError(t, err)

	impl, ok := client.(*CommandClient)
	require.True(t, ok)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		res, err := client.AllDeviceCoreCommands(context.Background(), 0, 20)

		require.NoError(t, err)
		require.IsType(t, res, responses.MultiDeviceCoreCommandsResponse{})
		require.Equal(t, res.RequestId, testRequestID)
	}()

	<-mockMessageBus.signal

	responseDTO := responses.NewMultiDeviceCoreCommandsResponse(testRequestID, "", http.StatusOK, 0, nil)
	responseBytes, err := json.Marshal(responseDTO)
	require.NoError(t, err)

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, testRequestID, testCorrelationID, common.ContentTypeJSON)
	require.NoError(t, err)

	impl.queryMessages <- responseEnvelope
	wg.Wait()
}

func TestCommandClient_DeviceCoreCommandsByDeviceName(t *testing.T) {
	mockMessageBus := mockMessageClient{signal: make(chan struct{})}
	topics := map[string]string{
		QueryRequestTopic:  testQueryRequestTopic,
		QueryResponseTopic: testQueryResponseTopic,
	}

	client, err := NewCommandClient(mockMessageBus, topics, 10*time.Second)
	require.NoError(t, err)

	impl, ok := client.(*CommandClient)
	require.True(t, ok)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		res, err := client.DeviceCoreCommandsByDeviceName(context.Background(), testDeviceName)

		require.NoError(t, err)
		require.IsType(t, res, responses.DeviceCoreCommandResponse{})
		require.Equal(t, res.RequestId, testRequestID)
	}()

	<-mockMessageBus.signal

	responseDTO := responses.NewDeviceCoreCommandResponse(testRequestID, "", http.StatusOK, dtos.DeviceCoreCommand{})
	responseBytes, err := json.Marshal(responseDTO)
	require.NoError(t, err)

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, testRequestID, testCorrelationID, common.ContentTypeJSON)
	require.NoError(t, err)

	impl.queryMessages <- responseEnvelope
	wg.Wait()
}

func TestCommandClient_IssueGetCommandByName(t *testing.T) {
	mockMessageBus := mockMessageClient{signal: make(chan struct{})}
	topics := map[string]string{
		CommandRequestTopicPrefix: testCommandRequestTopicPrefix,
		CommandResponseTopic:      testCommandResponseTopic,
	}

	client, err := NewCommandClient(mockMessageBus, topics, 10*time.Second)
	require.NoError(t, err)

	impl, ok := client.(*CommandClient)
	require.True(t, ok)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		res, err := client.IssueGetCommandByName(context.Background(), testDeviceName, testCommandName, "no", "tes")

		require.NoError(t, err)
		require.IsType(t, res, &responses.EventResponse{})
		require.Equal(t, res.RequestId, testRequestID)
	}()

	<-mockMessageBus.signal

	responseDTO := responses.NewEventResponse(testRequestID, "", http.StatusOK, dtos.Event{})
	responseBytes, err := json.Marshal(responseDTO)
	require.NoError(t, err)

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, testRequestID, testCorrelationID, common.ContentTypeJSON)
	require.NoError(t, err)

	impl.commandMessages <- responseEnvelope
	wg.Wait()
}

func TestCommandClient_IssueGetCommandByNameWithQueryParams(t *testing.T) {
	mockMessageBus := mockMessageClient{signal: make(chan struct{})}
	topics := map[string]string{
		CommandRequestTopicPrefix: testCommandRequestTopicPrefix,
		CommandResponseTopic:      testCommandResponseTopic,
	}

	client, err := NewCommandClient(mockMessageBus, topics, 10*time.Second)
	require.NoError(t, err)

	impl, ok := client.(*CommandClient)
	require.True(t, ok)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		res, err := client.IssueGetCommandByNameWithQueryParams(context.Background(), testDeviceName, testCommandName, nil)

		require.NoError(t, err)
		require.IsType(t, res, &responses.EventResponse{})
		require.Equal(t, res.RequestId, testRequestID)
	}()

	<-mockMessageBus.signal

	responseDTO := responses.NewEventResponse(testRequestID, "", http.StatusOK, dtos.Event{})
	responseBytes, err := json.Marshal(responseDTO)
	require.NoError(t, err)

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, testRequestID, testCorrelationID, common.ContentTypeJSON)
	require.NoError(t, err)

	impl.commandMessages <- responseEnvelope
	wg.Wait()
}

func TestCommandClient_IssueSetCommandByName(t *testing.T) {
	mockMessageBus := mockMessageClient{signal: make(chan struct{})}
	topics := map[string]string{
		CommandRequestTopicPrefix: testCommandRequestTopicPrefix,
		CommandResponseTopic:      testCommandResponseTopic,
	}

	client, err := NewCommandClient(mockMessageBus, topics, 10*time.Second)
	require.NoError(t, err)

	impl, ok := client.(*CommandClient)
	require.True(t, ok)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		res, err := client.IssueSetCommandByName(context.Background(), testDeviceName, testCommandName, nil)

		require.NoError(t, err)
		require.IsType(t, res, commonDTO.BaseResponse{})
		require.Equal(t, res.RequestId, testRequestID)
	}()

	<-mockMessageBus.signal

	responseDTO := commonDTO.NewBaseResponse(testRequestID, "", http.StatusOK)
	responseBytes, err := json.Marshal(responseDTO)
	require.NoError(t, err)

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, testRequestID, testCorrelationID, common.ContentTypeJSON)
	require.NoError(t, err)

	impl.commandMessages <- responseEnvelope
	wg.Wait()
}

func TestCommandClient_IssueSetCommandByNameWithObject(t *testing.T) {
	mockMessageBus := mockMessageClient{signal: make(chan struct{})}
	topics := map[string]string{
		CommandRequestTopicPrefix: testCommandRequestTopicPrefix,
		CommandResponseTopic:      testCommandResponseTopic,
	}

	client, err := NewCommandClient(mockMessageBus, topics, 10*time.Second)
	require.NoError(t, err)

	impl, ok := client.(*CommandClient)
	require.True(t, ok)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		res, err := client.IssueSetCommandByNameWithObject(context.Background(), testDeviceName, testCommandName, nil)

		require.NoError(t, err)
		require.IsType(t, res, commonDTO.BaseResponse{})
		require.Equal(t, res.RequestId, testRequestID)
	}()

	<-mockMessageBus.signal

	responseDTO := commonDTO.NewBaseResponse(testRequestID, "", http.StatusOK)
	responseBytes, err := json.Marshal(responseDTO)
	require.NoError(t, err)

	responseEnvelope, err := types.NewMessageEnvelopeForResponse(responseBytes, testRequestID, testCorrelationID, common.ContentTypeJSON)
	require.NoError(t, err)

	impl.commandMessages <- responseEnvelope
	wg.Wait()
}
