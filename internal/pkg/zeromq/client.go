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
	"time"

	"github.com/edgexfoundry/go-mod-messaging/pkg/messaging"

	zmq "github.com/pebbe/zmq4"
)

const (
	singleMessagePayloadIndex = iota
	payloadIndex
)

const (
	defaultMsgProtocol       = "tcp"
	maxZeroMqSubscribeTopics = 10
)

type zeromqClient struct {
	config      messaging.MessageBusConfig
	publisher   *zmq.Socket
	subscribers []zeromqSubscriber
	errors      chan error
}

// NewZeroMqClient instantiates a new zeromq client instance based on the configuration
func NewZeroMqClient(msgConfig messaging.MessageBusConfig) (*zeromqClient, error) {

	client := zeromqClient{config: msgConfig}

	return &client, nil
}

// Connect implements connect to 0mq
// Since 0mq pub-sub pattern has different pub socket type and sub socket one
// the socket initialization and connection are delayed to Publish and Subscribe calls, respectively
func (client *zeromqClient) Connect() error {
	return nil
}

func (client *zeromqClient) Publish(message messaging.MessageEnvelope, topic string) error {

	msgQueueURL := client.getPublishMessageQueueURL()
	var err error

	if client.publisher == nil {

		if client.publisher, err = zmq.NewSocket(zmq.PUB); err != nil {
			return err
		}
		if conErr := client.publisher.Bind(msgQueueURL); conErr != nil {
			// wrapping the error with msgQueueURL info:
			return fmt.Errorf("Error: %v [%s]", conErr, msgQueueURL)
		}

		// allow some time for socket binding before start publishing
		time.Sleep(time.Second)
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msgTotalBytes, err := client.publisher.SendMessage(topic, msgBytes)

	if err != nil {
		return err
	} else if msgTotalBytes != len(topic)+len(msgBytes) {
		return errors.New("The length of the sent messages does not match the expected length")
	}

	return err
}

func (client *zeromqClient) Subscribe(topics []messaging.TopicChannel, messageErrors chan error) error {

	if len(topics) == 0 {
		return errors.New("zero-length topics not allow")
	} else if len(topics) > maxZeroMqSubscribeTopics {
		return fmt.Errorf("Number of topics(%d) exceeds the maximum capacity(%d)", len(topics), maxZeroMqSubscribeTopics)
	}

	client.errors = messageErrors
	client.subscribers = make([]zeromqSubscriber, len(topics))
	var errorsSubscribe []error
	for idx, topic := range topics {
		subErr := client.subscribeTopic(&client.subscribers[idx], &topic)
		if subErr != nil {
			errorsSubscribe = append(errorsSubscribe, subErr)
		}
	}

	if len(errorsSubscribe) == 0 {
		return nil
	}

	var errorStr string
	for _, err := range errorsSubscribe {
		errorStr = errorStr + fmt.Sprintf("%s  ", err.Error())
	}
	return errors.New(errorStr)
}

func (client *zeromqClient) subscribeTopic(subscriber *zeromqSubscriber, aTopic *messaging.TopicChannel) error {

	msgQueueURL := client.getSubscribMessageQueueURL()
	if err := subscriber.init(msgQueueURL, aTopic); err != nil {
		return err
	}

	go func(topic *messaging.TopicChannel) {

		subscriber.connection.SetSubscribe(topic.Topic)
		for {

			payloadMsg, err := subscriber.connection.RecvMessage(0)

			if err != nil && err.Error() == "Context was terminated" {
				subscriber.connection.SetLinger(time.Duration(0))
				subscriber.connection.Close()
				return
			}

			if err != nil && err.Error() != "resource temporarily unavailable" {
				continue
			}

			var payloadBytes []byte
			switch msgLen := len(payloadMsg); msgLen {
			case 1:
				payloadBytes = []byte(payloadMsg[singleMessagePayloadIndex])
			case 2:
				payloadBytes = []byte(payloadMsg[payloadIndex])
			default:
				client.errors <- fmt.Errorf("Found more than 2 incoming messages (1 is no topic, 2 is topic and message), but found: %d", msgLen)
				continue
			}

			msgEnvelope := messaging.MessageEnvelope{}
			unmarshalErr := json.Unmarshal(payloadBytes, &msgEnvelope)
			if unmarshalErr != nil {
				client.errors <- unmarshalErr
				continue
			}

			topic.Messages <- &msgEnvelope
		}
	}(&subscriber.topic)

	return nil
}

func (client *zeromqClient) allSubscribersDisconnected() bool {
	allDisconnected := true
	for _, subscriber := range client.subscribers {
		if subscriber.connection.String() != "Socket(CLOSED)" {
			allDisconnected = false
			break
		}
	}
	return allDisconnected
}

func (client *zeromqClient) Disconnect() error {

	var disconnectErrs []error
	for _, subscriber := range client.subscribers {
		err := subscriber.context.Term()

		if err != nil {
			disconnectErrs = append(disconnectErrs, err)
		}
	}

	if client.publisher != nil {
		client.publisher.SetLinger(time.Duration(0))
		err := client.publisher.Close()
		if err != nil {
			disconnectErrs = append(disconnectErrs, err)
		}
	}

	// close error channel
	if client.errors != nil {
		close(client.errors)
	}

	// close all topic channels
	for _, subscriber := range client.subscribers {
		if subscriber.topic.Messages != nil {
			close(subscriber.topic.Messages)
		}
	}

	if len(disconnectErrs) == 0 {
		return nil
	}

	var errorStr string
	for _, err := range disconnectErrs {
		if err != nil {
			errorStr = errorStr + fmt.Sprintf("%s  ", err.Error())
		}
	}

	return errors.New(errorStr)
}

func (client *zeromqClient) getPublishMessageQueueURL() string {
	return fmt.Sprintf("%s://%s:%d", client.getPublishMessageProtocol(), client.config.PublishHost.Host,
		client.config.PublishHost.Port)
}

func (client *zeromqClient) getPublishMessageProtocol() string {
	if client.config.PublishHost.Protocol == "" {
		return defaultMsgProtocol
	}
	return client.config.PublishHost.Protocol
}

func (client *zeromqClient) getSubscribMessageQueueURL() string {
	return fmt.Sprintf("%s://%s:%d", client.getSubscribeMessageProtocol(), client.config.SubscribeHost.Host,
		client.config.SubscribeHost.Port)
}

func (client *zeromqClient) getSubscribeMessageProtocol() string {
	if client.config.SubscribeHost.Protocol == "" {
		return defaultMsgProtocol
	}
	return client.config.SubscribeHost.Protocol
}
