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

package mqtt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v2/messaging/mqtt"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

func TestCreateMQTTClientConfiguration(t *testing.T) {
	type args struct {
		messageBusConfig types.MessageBusConfig
	}
	tests := []struct {
		name    string
		args    args
		want    MQTTClientConfig
		wantErr bool
	}{
		{
			"Successfully load all configurations",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					mqtt.Username:       "TestUser",
					mqtt.Password:       "TestPassword",
					mqtt.ClientId:       "TestClientID",
					mqtt.Qos:            "1",
					mqtt.KeepAlive:      "3",
					mqtt.Retained:       "true",
					mqtt.ConnectTimeout: "7",
				}}},
			MQTTClientConfig{
				BrokerURL: "tcp://example.com:9090",
				MQTTClientOptions: MQTTClientOptions{
					Username:       "TestUser",
					Password:       "TestPassword",
					ClientId:       "TestClientID",
					Qos:            1,
					KeepAlive:      3,
					Retained:       true,
					ConnectTimeout: 7,
				},
			},
			false,
		},
		{
			"Does not over write host configuration with optional properties",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					mqtt.Username:       "TestUser",
					mqtt.Password:       "TestPassword",
					mqtt.ClientId:       "TestClientID",
					mqtt.Qos:            "1",
					mqtt.KeepAlive:      "3",
					mqtt.Retained:       "true",
					mqtt.ConnectTimeout: "7",
				}}},
			MQTTClientConfig{
				BrokerURL: "tcp://example.com:9090",
				MQTTClientOptions: MQTTClientOptions{
					Username:       "TestUser",
					Password:       "TestPassword",
					ClientId:       "TestClientID",
					Qos:            1,
					KeepAlive:      3,
					Retained:       true,
					ConnectTimeout: 7,
				}},
			false,
		},
		{
			"Invalid URL",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "   ", Port: 999999999999, Protocol: "    "},
				Optional: map[string]string{
					// Other valid configurations
					"ClientId": "TestClientID",
				}}},
			MQTTClientConfig{},
			true,
		},
		{
			"Invalid Int",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					"KeepAlive": "abc",
					// Other valid configurations
				}}},
			MQTTClientConfig{},
			true,
		},
		{
			"Invalid Bool",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					"Retained": "abc",
				}}},
			MQTTClientConfig{},
			true,
		},
		{
			"Unknown configuration",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					mqtt.Username:       "TestUser",
					mqtt.Password:       "TestPassword",
					mqtt.ClientId:       "TestClientID",
					mqtt.Qos:            "1",
					mqtt.KeepAlive:      "3",
					mqtt.Retained:       "true",
					mqtt.ConnectTimeout: "7",
					"Unknown config":    "Something random",
				}}},
			MQTTClientConfig{
				BrokerURL: "tcp://example.com:9090",
				MQTTClientOptions: MQTTClientOptions{
					Username:       "TestUser",
					Password:       "TestPassword",
					ClientId:       "TestClientID",
					Qos:            1,
					KeepAlive:      3,
					Retained:       true,
					ConnectTimeout: 7,
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateMQTTClientConfiguration(tt.args.messageBusConfig)
			if tt.wantErr {
				require.Error(t, err, "CreateMQTTClientConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return // End test for expected errors
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
