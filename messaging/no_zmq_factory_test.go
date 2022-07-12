//go:build windows || no_zmq

//
// Copyright (c) 2021 Intel Corporation
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

func TestNewMessageClientZeroMq_Windows(t *testing.T) {
	expected := "zmq is no longer supported on windows native deployments or binary built without it"
	config := types.MessageBusConfig{
		Type: ZeroMQ,
		PublishHost: types.HostInfo{
			Host:     "*",
			Port:     5570,
			Protocol: "tcp",
		},
	}

	_, err := NewMessageClient(config)
	require.Error(t, err)
	assert.Equal(t, expected, err.Error())
}
