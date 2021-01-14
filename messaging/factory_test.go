//
// Copyright (c) 2019 Intel Corporation
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

package messaging

import (
	"testing"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/stretchr/testify/assert"
)

var msgConfig = types.MessageBusConfig{
	PublishHost: types.HostInfo{
		Host:     "*",
		Port:     5570,
		Protocol: "tcp",
	},
}

func TestNewMessageClientZeroMq(t *testing.T) {

	msgConfig.Type = ZeroMQ
	_, err := NewMessageClient(msgConfig)

	if assert.NoError(t, err, "New Message client failed: ", err) == false {
		t.Fatal()
	}
}

func TestNewMessageClientMQTT(t *testing.T) {
	messageBusConfig := msgConfig
	messageBusConfig.Type = MQTT
	messageBusConfig.Optional = map[string]string{
		"Username":          "TestUser",
		"Password":          "TestPassword",
		"ClientId":          "TestClientID",
		"Topic":             "TestTopic",
		"Qos":               "1",
		"KeepAlive":         "3",
		"Retained":          "true",
		"ConnectionPayload": "TestConnectionPayload",
	}

	_, err := NewMessageClient(messageBusConfig)

	if assert.NoError(t, err, "New Message client failed: ", err) == false {
		t.Fatal()
	}
}

func TestNewMessageBogusType(t *testing.T) {

	msgConfig.Type = "bogus"

	_, err := NewMessageClient(msgConfig)
	if assert.Error(t, err, "Expected message type error") == false {
		t.Fatal()
	}
}

func TestNewMessageEmptyHostAndPortNumber(t *testing.T) {

	msgConfig.PublishHost.Host = ""
	msgConfig.PublishHost.Port = 0
	_, err := NewMessageClient(msgConfig)
	if assert.Error(t, err, "Expected message type error") == false {
		t.Fatal()
	}

}
