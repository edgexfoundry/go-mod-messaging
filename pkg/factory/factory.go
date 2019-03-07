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
	"fmt"
	"strings"

	messaging "github.com/edgexfoundry/go-mod-messaging"
)

const (
	// ZeroMQ messaging implementation
	ZeroMQ = "zero"
)

// NewMessageClient is a factory function to instantiate different message client depending on
// the "Type" from the configuration
func NewMessageClient(msgConfig messaging.MessageBusConfig) (msgClient messaging.MessageClient, createErr error) {

	if isHostInfoEmpty(&msgConfig.PublishHost) {
		return nil, fmt.Errorf("unable to create messageClient: message publisher host and/or port not set")
	}

	switch lowerMsgType := strings.ToLower(msgConfig.Type); lowerMsgType {
	case ZeroMQ:
		// TBD
		return msgClient, createErr
	default:
		return nil, fmt.Errorf("unknown message type '%s' requested", msgConfig.Type)
	}
}

func isHostInfoEmpty(hostInfo *messaging.HostInfo) bool {
	return hostInfo == nil || hostInfo.Host == "" || hostInfo.Port == 0
}
