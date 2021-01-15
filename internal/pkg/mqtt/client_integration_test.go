// +build mqttIntegration

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
 * - The MQTT server URL set as the environment variable MQTT_SERVER_TEST. Otherwise the default 'tcp://localhost:1833
 *  will be used
 */

package mqtt

import (
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v2/messaging/mqtt"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

const (
	MQTTURLEnvName = "MQTT_SERVER_TEST"
	DefaultMQTTURL = "tcp://localhost:1883"
)

// TestIntegrationWithMQTTServer end-to-end test of the MQTT client with a MQTT server.
func TestIntegrationWithMQTTServer(t *testing.T) {
	mqttURL := os.Getenv(MQTTURLEnvName)
	if mqttURL == "" {
		mqttURL = DefaultMQTTURL
	}

	urlMQTT, err := url.Parse(mqttURL)
	require.NoError(t, err, "Failed to create URL")
	port, err := strconv.ParseInt(urlMQTT.Port(), 10, 0)
	require.NoError(t, err, "Unable to parse the port")
	configOptions := types.MessageBusConfig{
		PublishHost: types.HostInfo{
			Host:     urlMQTT.Hostname(),
			Port:     int(port),
			Protocol: urlMQTT.Scheme,
		},
		Optional: map[string]string{
			mqtt.ClientId: "integration-test-client",
		},
	}

	client, err := NewMQTTClient(configOptions)
	require.NoError(t, err, "Failed to create MQTT client")
	err = client.Connect()
	defer func() { _ = client.Disconnect() }()
	require.NoError(t, err, "Failed to connect")
	channel := make(chan types.MessageEnvelope)
	topics := []types.TopicChannel{{
		Topic:    "test1",
		Messages: channel,
	}}

	err = client.Subscribe(topics, make(chan error))
	require.NoError(t, err, "Failed to create subscription")
	expectedMessage := types.MessageEnvelope{
		Checksum:      "123",
		CorrelationID: "456",
		Payload:       []byte("Testing the MQTT client"),
		ContentType:   "application/text",
	}

	err = client.Publish(expectedMessage, "test1")
	require.NoError(t, err, "Failed to publish message")
	actualMessage := <-channel
	assert.Equal(t, expectedMessage, actualMessage)
}
