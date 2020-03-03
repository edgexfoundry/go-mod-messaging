// +build integration

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
 * `integration` tag when running the tests. For example 'go test ./... -tags=integration`.
 *
 * This is a list of the requirements necessary to run the tests in this file:
 * - MQTT server with no authentication required by clients.
 * - The MQTT server URL set as the environment variable MQTT_SERVER_TEST. Otherwise the default 'tcp://localhost:1833
 *  will be used
 */

package mqtt

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/edgexfoundry/go-mod-messaging/messaging/mqtt"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
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
	if err != nil {
		t.Error("Failed to create URL:" + err.Error())
		return
	}

	port, err := strconv.ParseInt(urlMQTT.Port(), 10, 0)
	if err != nil {
		fmt.Errorf("Unable to parse the port:%s ", err.Error())
	}
	configOptions := types.MessageBusConfig{
		PublishHost: types.HostInfo{
			Host:     urlMQTT.Hostname(),
			Port:     int(port),
			Protocol: urlMQTT.Scheme,
		},
		Optional: map[string]string{
			mqtt.ClientId:          "integration-test-client",
			mqtt.Username:          "",
			mqtt.Password:          "",
			mqtt.Topic:             "test1",
			mqtt.Qos:               "0",
			mqtt.KeepAlive:         "5",
			mqtt.Retained:          "false",
			mqtt.ConnectionPayload: "",
		},
	}

	client, err := NewMQTTClient(configOptions)
	if err != nil {
		t.Error("Failed to create MQTT client: " + err.Error())
	}

	err = client.Connect()
	defer client.Disconnect()
	if err != nil {
		t.Error("Failed to connect:" + err.Error())
		return
	}

	channel := make(chan types.MessageEnvelope)
	topics := []types.TopicChannel{{
		Topic:    "test1",
		Messages: channel,
	}}

	err = client.Subscribe(topics, make(chan error))
	if err != nil {
		t.Error("Failed to create subscription: " + err.Error())
		return
	}

	expectedMessage := types.MessageEnvelope{
		Checksum:      "123",
		CorrelationID: "456",
		Payload:       []byte("Testing the MQTT client"),
		ContentType:   "application/text",
	}

	err = client.Publish(expectedMessage, "test1")
	if err != nil {
		t.Error("Failed to publish: " + err.Error())
		return
	}

	actualMessage := <-channel
	if !reflect.DeepEqual(expectedMessage, actualMessage) {
		t.Errorf("Expected message of: %v , but got: %v", expectedMessage, actualMessage)
	}
}
