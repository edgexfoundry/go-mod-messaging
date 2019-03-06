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
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	messaging "github.com/edgexfoundry/go-mod-messaging"
	zmq "github.com/pebbe/zmq4"
)

const (
	defaultMsgProtocol = "tcp"
)

type zeromqClient struct {
	publishSocket   *zmq.Socket
	subscribeSocket *zmq.Socket
	publishMux      sync.Mutex
	subscribeMux    sync.Mutex
	topicFilters    []string
	config          messaging.MessageBusConfig
}

// NewZeroMqClient instantiates a new zeromq client instance based on the configuration
func NewZeroMqClient(msgConfig messaging.MessageBusConfig) (*zeromqClient, error) {

	client := zeromqClient{config: msgConfig}
	return &client, nil
}

func (client *zeromqClient) Connect() error {
	return nil
}

func (client *zeromqClient) Publish(message messaging.MessageEnvelope, topic string) error {

	msgQueueURL := getMessageQueueURL(&client.config)
	var err error

	if client.publishSocket == nil {

		client.publishSocket, err = zmq.NewSocket(zmq.PUB)

		if err != nil {
			return err
		}
		if conErr := client.publishSocket.Bind(msgQueueURL); conErr != nil {

			return conErr
		}

		fmt.Println("Successfully connected to 0MQ message queue")
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	client.publishMux.Lock()
	defer client.publishMux.Unlock()

	lenOfTopic, err := client.publishSocket.Send(topic, zmq.SNDMORE)

	if err != nil {
		return err
	} else if lenOfTopic != len(topic) {
		return errors.New("The length of the sent topic does not match the expected length")
	}

	lenOfPayload, err := client.publishSocket.SendBytes(msgBytes, zmq.DONTWAIT)

	if lenOfPayload != len(msgBytes) {
		return errors.New("The length of the sent payload does not match the expected length")
	}

	return err
}

func (client *zeromqClient) Subscribe(topics []messaging.TopicChannel, host string, messageErrors chan error) error {

	var err error
	if client.subscribeSocket == nil {
		client.subscribeSocket, err = zmq.NewSocket(zmq.SUB)

		if err != nil {
			return err
		}
	}

	subscribeConfig := messaging.MessageBusConfig{Host: host, Port: client.config.Port, Protocol: client.config.Protocol}
	msgQueueURL := getMessageQueueURL(&subscribeConfig)

	if err = client.subscribeSocket.Connect(msgQueueURL); err != nil {
		return err
	}

	go func(msgQueueURL string) {

		for {
			for _, topic := range topics {
				client.subscribeSocket.SetSubscribe(topic.Topic)

				data, err := client.subscribeSocket.Recv(zmq.DONTWAIT)

				if err != nil {
					messageErrors <- err
				}
				topic.Messages <- data
			}
			time.Sleep(10)
		}
	}(msgQueueURL)
	return nil
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
