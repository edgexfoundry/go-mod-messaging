//go:build mqttIntegration

/********************************************************************************
 *  Copyright 2020 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

/**
 * This test file contains integration tests for the MQTT Client. These integration tests are disabled by default as
 * they require additional external resources to be executed properly. These tests are enabled by providing the
 * `mqttIntegration` tag when running the tests. For example 'go test ./... -tags=mqttIntegration`.
 *
 * This is a list of the requirements necessary to run the tests in this file:
 * - MQTT server with no authentication required by clients.
 * - The MQTT server URL set as the environment variable MQTT_SERVER_TEST, Otherwise the default 'tcp://localhost:1833
 *  will be used
 */

package mqtt

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v3/internal/pkg"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
)

// TestIntegrationWithMQTTServer end-to-end test of the MQTT client with a MQTT server.
func TestIntegrationWithMQTTServer(t *testing.T) {
	client := createAndConnectMqttClient(t)
	defer func() { _ = client.Disconnect() }()

	channel := make(chan types.MessageEnvelope)
	topics := []types.TopicChannel{{
		Topic:    "test1",
		Messages: channel,
	}}

	err := client.Subscribe(topics, make(chan error))
	require.NoError(t, err, "Failed to create subscription")
	expectedMessage := types.MessageEnvelope{
		CorrelationID: "456",
		Payload:       []byte("Testing the MQTT client"),
		ContentType:   "application/text",
		ReceivedTopic: "test1",
	}

	err = client.Publish(expectedMessage, "test1")
	require.NoError(t, err, "Failed to publish message")
	actualMessage := <-channel
	assert.Equal(t, expectedMessage, actualMessage)
}

// TestUnsubscribeIntegrationWithMQTTServer end-to-end test subscribing and unsubscribing.
// MQTT broker must be running with Device virtual publishing events.
func TestUnsubscribeIntegrationWithMQTTServer(t *testing.T) {
	client := createAndConnectMqttClient(t)
	defer func() { _ = client.Disconnect() }()

	messages := make(chan types.MessageEnvelope, 1)
	errs := make(chan error, 1)

	eventTopic := "edgex/events/#"
	topics := []types.TopicChannel{
		{
			Topic:    eventTopic,
			Messages: messages,
		},
	}

	println("Subscribing to topic: " + eventTopic)
	err := client.Subscribe(topics, errs)
	require.NoError(t, err)
	require.Equal(t, 1, len(client.existingSubscriptions))

	messageCount := 0

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(time.Second * 10):
				return

			case err = <-errs:
				require.Failf(t, "failed", "Unexpected error message received: %v", err)
				return

			case message := <-messages:
				println(fmt.Sprintf("Received message from topic: %v", message.ReceivedTopic))
				messageCount++
				if messageCount > 3 {
					println("Unsubscribing from topic: " + eventTopic)
					err = client.Unsubscribe(eventTopic)
					require.NoError(t, err)
					require.Equal(t, 0, len(client.existingSubscriptions))
				}
			}
		}

	}()

	wg.Wait()
	assert.Greater(t, messageCount, 3)
}

// TestUnsubscribeIntegrationWithMQTTServer end-to-end test subscribing and unsubscribing.
// MQTT broker must be running with Device virtual publishing events.
func TestRequestIntegrationWithMQTTServer(t *testing.T) {
	client := createAndConnectMqttClient(t)

	defer func() { _ = client.Disconnect() }()

	messages := make(chan types.MessageEnvelope, 1)
	errs := make(chan error, 1)

	responseTopicPrefix := "/edgex/response/test-service"
	requestTopic := "edgex/request"
	topics := []types.TopicChannel{
		{
			Topic:    requestTopic,
			Messages: messages,
		},
	}

	println("Subscribing to topic: " + requestTopic)
	err := client.Subscribe(topics, errs)
	require.NoError(t, err)
	require.Equal(t, 1, len(client.existingSubscriptions))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(time.Second * 10):
				return

			case err = <-errs:
				require.Failf(t, "failed", "Unexpected error message received: %v", err)
				return

			case message := <-messages:
				println(fmt.Sprintf("Received message from topic: %v", message.ReceivedTopic))

				responseTopic := strings.Join([]string{responseTopicPrefix, message.RequestID}, "/")
				println(fmt.Sprintf("Publishing response message on topic: %v", responseTopic))

				err = client.Publish(types.MessageEnvelope{RequestID: message.RequestID}, responseTopic)
				require.NoError(t, err)
				return
			}
		}
	}()

	time.Sleep(time.Second)
	requestId := uuid.NewString()
	println(fmt.Sprintf("Sending request to topic %s with requestId %s", requestTopic, requestId))

	response, err := client.Request(types.MessageEnvelope{RequestID: requestId}, requestTopic, responseTopicPrefix, time.Second*10)
	require.NoError(t, err)
	assert.Equal(t, requestId, response.RequestID)
}

func createAndConnectMqttClient(t *testing.T) *Client {
	client, err := NewMQTTClient(types.MessageBusConfig{
		Broker: types.HostInfo{
			Host:     "localhost",
			Port:     1883,
			Protocol: "tcp",
		},
		Optional: map[string]string{
			pkg.ClientId:  "test",
			pkg.KeepAlive: "10",
		},
	})

	require.NoError(t, err)

	err = client.Connect()
	require.NoError(t, err, "Failed to connect")

	return client
}
