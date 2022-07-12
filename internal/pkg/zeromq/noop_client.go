//go:build windows || no_zmq

//
// Copyright (c) 2022 Intel Corporation
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

package zeromq

import (
	"errors"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

var notSupportedErrorMessage = errors.New("zmq is no longer supported on windows native deployments or binary built without it")

type zeromqClient struct {
}

func NewZeroMqClient(_ types.MessageBusConfig) (*zeromqClient, error) {
	return nil, notSupportedErrorMessage
}

func (client *zeromqClient) Connect() error {
	return notSupportedErrorMessage
}

func (client *zeromqClient) Publish(_ types.MessageEnvelope, _ string) error {
	return notSupportedErrorMessage

}

func (client *zeromqClient) Subscribe(_ []types.TopicChannel, _ chan error) error {
	return notSupportedErrorMessage
}

func (client *zeromqClient) Disconnect() error {
	return notSupportedErrorMessage
}
