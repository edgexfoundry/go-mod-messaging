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

//go:build !include_nats_messaging

package messaging_test

import (
	"testing"

	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var natsConfig = types.MessageBusConfig{
	PublishHost: types.HostInfo{
		Host:     "*",
		Port:     4222,
		Protocol: "tcp",
	},
}

func TestNewMessageClientNatsCore(t *testing.T) {
	messageBusConfig := natsConfig
	messageBusConfig.Type = messaging.NatsCore
	messageBusConfig.SubscribeHost = types.HostInfo{Host: uuid.NewString(), Port: 37, Protocol: "nats"}

	_, err := messaging.NewMessageClient(messageBusConfig)

	require.Error(t, err)
}

func TestNewMessageClientNatsJetstream(t *testing.T) {
	messageBusConfig := natsConfig
	messageBusConfig.Type = messaging.NatsJetStream
	messageBusConfig.SubscribeHost = types.HostInfo{Host: uuid.NewString(), Port: 37, Protocol: "nats"}

	_, err := messaging.NewMessageClient(messageBusConfig)

	require.Error(t, err)
}
