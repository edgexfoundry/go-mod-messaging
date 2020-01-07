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
	"reflect"
	"testing"

	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
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
					"Username":          "TestUser",
					"Password":          "TestPassword",
					"ClientId":          "TestClientID",
					"Topic":             "TestTopic",
					"Qos":               "1",
					"KeepAlive":         "3",
					"Retained":          "true",
					"ConnectionPayload": "TestConnectionPayload",
				}}},
			MQTTClientConfig{
				BrokerURL: "tcp://example.com:9090",
				MQTTClientOptions: MQTTClientOptions{
					Username:          "TestUser",
					Password:          "TestPassword",
					ClientId:          "TestClientID",
					Topic:             "TestTopic",
					Qos:               1,
					KeepAlive:         3,
					Retained:          true,
					ConnectionPayload: "TestConnectionPayload",
				},
			},
			false,
		},
		{
			"Does not over write host configuration with optional properties",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					"BrokerURL":         "http://fail.edu",
					"Username":          "TestUser",
					"Password":          "TestPassword",
					"ClientId":          "TestClientID",
					"Topic":             "TestTopic",
					"Qos":               "1",
					"KeepAlive":         "3",
					"Retained":          "true",
					"ConnectionPayload": "TestConnectionPayload",
				}}},
			MQTTClientConfig{
				BrokerURL: "tcp://example.com:9090",
				MQTTClientOptions: MQTTClientOptions{
					Username:          "TestUser",
					Password:          "TestPassword",
					ClientId:          "TestClientID",
					Topic:             "TestTopic",
					Qos:               1,
					KeepAlive:         3,
					Retained:          true,
					ConnectionPayload: "TestConnectionPayload",
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
					"Username":          "TestUser",
					"Password":          "TestPassword",
					"ClientId":          "TestClientID",
					"Topic":             "TestTopic",
					"Qos":               "1",
					"KeepAlive":         "3",
					"Retained":          "true",
					"ConnectionPayload": "TestConnectionPayload",
					"Unknown config":    "Something random",
				}}},
			MQTTClientConfig{
				BrokerURL: "tcp://example.com:9090",
				MQTTClientOptions: MQTTClientOptions{
					Username:          "TestUser",
					Password:          "TestPassword",
					ClientId:          "TestClientID",
					Topic:             "TestTopic",
					Qos:               1,
					KeepAlive:         3,
					Retained:          true,
					ConnectionPayload: "TestConnectionPayload",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateMQTTClientConfiguration(tt.args.messageBusConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateMQTTClientConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateMQTTClientConfiguration() got = %v, want %v", got, tt.want)
			}
		})
	}
}
