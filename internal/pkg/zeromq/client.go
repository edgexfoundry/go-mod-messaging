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

package zeromq

import (
	"fmt"
	"sync"

	messaging "github.com/edgexfoundry/go-mod-messaging"
	zmq "github.com/pebbe/zmq4"
)

const (
	defaultMsgProtocol = "tcp"
)

type zeroMQEventPublisher struct {
	publisher *zmq.Socket
	mux       sync.Mutex
}

type zeroMQEventSubscriber struct {
	subsriber    *zmq.Socket
	mux          sync.Mutex
	topicFilters []string
}

type zeromqClient struct {
	evtPublisher *zeroMQEventPublisher
	evtSubsriber *zeroMQEventSubscriber
}

// NewZeroMqClient instantiates a new zeromq client instance based on the configuration
func NewZeroMqClient(msgConfig messaging.MessageBusConfig) (*zeromqClient, error) {
	newMsgPublisher, newErr := zmq.NewSocket(zmq.PUB)
	if newErr != nil {
		return nil, newErr
	}

	msgQueueURL := getMessageQueueURL(&msgConfig)
	fmt.Println("Connecting to 0MQ message queue at: " + msgQueueURL)

	if conErr := newMsgPublisher.Bind(msgQueueURL); conErr != nil {
		return nil, conErr
	}

	fmt.Println("Successfully connected to 0MQ message queue")
	sender := &zeroMQEventPublisher{
		publisher: newMsgPublisher,
	}

	newSubscriber, _ := zmq.NewSocket(zmq.SUB)

	receiver := &zeroMQEventSubscriber{
		subsriber: newSubscriber,
	}

	client := &zeromqClient{evtPublisher: sender, evtSubsriber: receiver}
	return client, nil
}

func getMessageQueueURL(msgConfig *messaging.MessageBusConfig) string {
	return fmt.Sprintf("%s://%s:%d", getMessageProtocol(msgConfig), msgConfig.Host, msgConfig.Port)
}

func getMessageProtocol(msgConfig *messaging.MessageBusConfig) string {
	if msgConfig.Protocol == "" {
		return defaultMsgProtocol
	}
	return msgConfig.Protocol
}
