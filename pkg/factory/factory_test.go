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

package factory

import (
	"testing"

	messaging "github.com/edgexfoundry/go-mod-messaging"
	"github.com/stretchr/testify/assert"
)

var msgConfig = messaging.MessageBusConfig{
	PublishHost: messaging.HostInfo{
		Host:     "*",
		Port:     5563,
		Protocol: "tcp",
	},
}

func TestNewMessageClientZeroMq(t *testing.T) {

	msgConfig.Type = "zero"

	_, err := NewMessageClient(msgConfig)

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
