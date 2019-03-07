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
	publisher    *zmq.Socket
	subscriber   *zmq.Socket
	publishMux   sync.Mutex
	subscribeMux sync.Mutex
	topics       []messaging.TopicChannel
	errors       chan error
	config       messaging.MessageBusConfig
}

// NewZeroMqClient instantiates a new zeromq client instance based on the configuration
func NewZeroMqClient(msgConfig messaging.MessageBusConfig) (*zeromqClient, error) {

	client := zeromqClient{config: msgConfig}
	return &client, nil
}

// Connect implements connect to 0mq
// Since 0mq pub-sub pattern has different pub socket type and sub socket one
// the socket initialzation and connection are delayed to Publish and Subscribe calls, respectively
func (client *zeromqClient) Connect() error {
	return nil
}

func (client *zeromqClient) Publish(message messaging.MessageEnvelope, topic string) error {

	msgQueueURL := getMessageQueueURL(&client.config.PublishHost)
	var err error

	if client.publisher == nil {

		if client.publisher, err = zmq.NewSocket(zmq.PUB); err != nil {
			return err
		}
		if conErr := client.publisher.Bind(msgQueueURL); conErr != nil {
			return conErr
		}

		fmt.Println("Publisher successfully connected to 0MQ message queue")

		// allow some time for socket binding before start publishing
		time.Sleep(time.Second)
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	client.publishMux.Lock()
	defer client.publishMux.Unlock()

	lenOfTopic, err := client.publisher.Send(topic, zmq.SNDMORE)

	if err != nil {
		return err
	} else if lenOfTopic != len(topic) {
		return errors.New("The length of the sent topic does not match the expected length")
	}

	lenOfPayload, err := client.publisher.SendBytes(msgBytes, zmq.DONTWAIT)

	if lenOfPayload != len(msgBytes) {
		return errors.New("The length of the sent payload does not match the expected length")
	}

	return err
}

func (client *zeromqClient) Subscribe(topics []messaging.TopicChannel, messageErrors chan error) error {

	client.topics = topics
	client.errors = messageErrors

	msgQueueURL := getMessageQueueURL(&client.config.SubscribeHost)
	if err := client.initSubscriber(msgQueueURL); err != nil {
		return err
	}

	for _, topic := range topics {
		client.subscriber.SetSubscribe(topic.Topic)

		go func(topic messaging.TopicChannel) {
			for {
				msgTopic, err := client.subscriber.Recv(zmq.SNDMORE)
				fmt.Printf("Message topic: %s\n", msgTopic)

				payloadMsg, err := client.subscriber.Recv(0)

				if err != nil && err.Error() != "resource temporarily unavailable" {
					client.errors <- err
				}
				topic.Messages <- payloadMsg
			}
		}(topic)
	}
	return nil
}

func (client *zeromqClient) Disconnect() error {
	// close error channel
	if client.errors != nil {
		close(client.errors)
	}

	// close all topic channels
	for _, topic := range client.topics {
		if topic.Messages != nil {
			close(topic.Messages)
		}
	}

	var closeErrs []error
	// close sockets:
	if client.publisher != nil {
		errPublish := client.publisher.Close()
		client.publisher = nil
		if errPublish != nil {
			closeErrs = append(closeErrs, errPublish)
		}
	}
	if client.subscriber != nil {
		errSubscribe := client.subscriber.Close()
		client.subscriber = nil
		if errSubscribe != nil {
			closeErrs = append(closeErrs, errSubscribe)
		}
	}

	if len(closeErrs) == 0 {
		return nil
	}

	var errorStr string
	for _, err := range closeErrs {
		if err != nil {
			errorStr = errorStr + fmt.Sprintf("%s  ", err.Error())
		}
	}

	return errors.New(errorStr)
}

func (client *zeromqClient) initSubscriber(msgQueueURL string) (err error) {

	if client.subscriber == nil {
		if client.subscriber, err = zmq.NewSocket(zmq.SUB); err != nil {
			return err
		}
	}

	fmt.Printf("Subscribing to message queue: [%s] ...", msgQueueURL)
	fmt.Println()
	return client.subscriber.Connect(msgQueueURL)
}

func getMessageQueueURL(hostInfo *messaging.HostInfo) string {
	return fmt.Sprintf("%s://%s:%d", getMessageProtocol(hostInfo), hostInfo.Host, hostInfo.Port)
}

func getMessageProtocol(hostInfo *messaging.HostInfo) string {
	if hostInfo.Protocol == "" {
		return defaultMsgProtocol
	}
	return hostInfo.Protocol
}
