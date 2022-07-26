//
// Copyright (c) 2022 One Track Consulting
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

//go:build include_nats_messaging

package nats

import (
	"testing"

	"github.com/edgexfoundry/go-mod-messaging/v2/internal/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

func TestCreateClientConfiguration(t *testing.T) {
	type args struct {
		messageBusConfig types.MessageBusConfig
	}
	tests := []struct {
		name    string
		args    args
		want    ClientConfig
		wantErr bool
	}{
		{
			"Successfully load all configurations",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					pkg.Username:                "TestUser",
					pkg.Password:                "TestPassword",
					pkg.ClientId:                "TestClientID",
					pkg.Durable:                 "durable-1",
					pkg.Format:                  "format",
					pkg.AutoProvision:           "true",
					pkg.RetryOnFailedConnect:    "true",
					pkg.ConnectTimeout:          "7",
					pkg.QueueGroup:              "group-1",
					pkg.DefaultPubRetryAttempts: "12",
					pkg.Deliver:                 "set-deliver",
				}}},
			ClientConfig{
				BrokerURL: "tcp://example.com:9090",
				ClientOptions: ClientOptions{
					Username:                "TestUser",
					Password:                "TestPassword",
					ClientId:                "TestClientID",
					AutoProvision:           true,
					RetryOnFailedConnect:    true,
					Durable:                 "durable-1",
					Format:                  "format",
					ConnectTimeout:          7,
					QueueGroup:              "group-1",
					DefaultPubRetryAttempts: 12,
					Deliver:                 "set-deliver",
				},
			},
			false,
		},
		{
			"Does not over write host configuration with optional properties",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					pkg.Username:                "TestUser",
					pkg.Password:                "TestPassword",
					pkg.ClientId:                "TestClientID",
					pkg.KeepAlive:               "3",
					pkg.Retained:                "true",
					pkg.CleanSession:            "false",
					pkg.ConnectTimeout:          "7",
					pkg.DefaultPubRetryAttempts: "2",
				}}},
			ClientConfig{
				BrokerURL: "tcp://example.com:9090",
				ClientOptions: ClientOptions{
					Username:                "TestUser",
					Password:                "TestPassword",
					ClientId:                "TestClientID",
					ConnectTimeout:          7,
					DefaultPubRetryAttempts: 2,
					Format:                  "nats",
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
			ClientConfig{},
			true,
		},
		{
			"Invalid Int",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					pkg.ConnectTimeout: "abc",
					// Other valid configurations
				}}},
			ClientConfig{},
			true,
		},
		{
			"Invalid Bool",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					pkg.RetryOnFailedConnect: "abc",
				}}},
			ClientConfig{},
			true,
		},
		{
			"Unknown configuration",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					pkg.Username:       "TestUser",
					pkg.Password:       "TestPassword",
					pkg.ClientId:       "TestClientID",
					pkg.ConnectTimeout: "7",
					"Unknown config":   "Something random",
				}}},
			ClientConfig{
				BrokerURL: "tcp://example.com:9090",
				ClientOptions: ClientOptions{
					Username:                "TestUser",
					Password:                "TestPassword",
					ClientId:                "TestClientID",
					ConnectTimeout:          7,
					DefaultPubRetryAttempts: 2,
					Format:                  "nats",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateClientConfiguration(tt.args.messageBusConfig)
			if tt.wantErr {
				require.Error(t, err, "CreateClientConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return // End test for expected errors
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
