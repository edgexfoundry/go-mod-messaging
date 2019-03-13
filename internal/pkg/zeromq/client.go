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
	topicIndex = iota
	payloadIndex
)

const (
	defaultMsgProtocol = "tcp"
)

type zeromqClient struct {
	config     messaging.MessageBusConfig
	publisher  *zmq.Socket
	subscriber *zmq.Socket
	topics     []messaging.TopicChannel
	errors     chan error
	context    *zmq.Context
}

// NewZeroMqClient instantiates a new zeromq client instance based on the configuration
func NewZeroMqClient(msgConfig messaging.MessageBusConfig) (*zeromqClient, error) {

	ctx, _ := zmq.NewContext()
	client := zeromqClient{config: msgConfig, context: ctx}

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

		if client.publisher, err = client.context.NewSocket(zmq.PUB); err != nil {
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

	msgTotalBytes, err := client.publisher.SendMessage(topic, msgBytes)

	if err != nil {
		return err
	} else if msgTotalBytes != len(topic)+len(msgBytes) {
		return errors.New("The length of the sent messages does not match the expected length")
	}

	return err
}

func (client *zeromqClient) Subscribe(topics []messaging.TopicChannel, messageErrors chan error) error {

	client.topics = topics
	client.errors = messageErrors

	msgQueueURL := client.getSubscribMessageQueueURL()
	if err := client.initSubscriber(msgQueueURL); err != nil {
		return err
	}

	for _, topic := range topics {
		go func(topic messaging.TopicChannel) {

			client.subscriber.SetSubscribe(topic.Topic)
			for {
				select {
				default:
					payloadMsg, err := client.subscriber.RecvMessage(0)

					if err != nil && err.Error() == "Context was terminated" {
						fmt.Println("Disconnecting and closing socket")

						client.subscriber.SetLinger(time.Duration(0))
						client.subscriber.Close()

						if client.publisher != nil {
							client.publisher.SetLinger(time.Duration(0))
							defer client.publisher.Close()
						}
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
						return
					}

					if err != nil && err.Error() != "resource temporarily unavailable" {
						fmt.Printf("Error received from subscribe: %s", err)
						fmt.Println()
						client.errors <- err
						continue
					}

					var payloadBytes []byte
					switch msgLen := len(payloadMsg); msgLen {
					case 0:
						time.Sleep(10 * time.Millisecond)
						continue
					case 1:
						payloadBytes = []byte(payloadMsg[topicIndex])
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
			}
		}(topic)
	}

	return nil
}

func (client *zeromqClient) Disconnect() error {

	if client.subscriber != nil {
		fmt.Println("Terminating the context")
		// this will allow the subscriber socket to stop blocking so it can close itself
		return client.context.Term()
	} else if client.publisher != nil {
		client.publisher.SetLinger(time.Duration(0))
		return client.publisher.Close()
	}
	return nil
}

func (client *zeromqClient) initSubscriber(msgQueueURL string) (err error) {

	if client.subscriber == nil {
		if client.subscriber, err = client.context.NewSocket(zmq.SUB); err != nil {
			return err
		}
	}

	fmt.Printf("Subscribing to message queue: [%s] ...", msgQueueURL)
	fmt.Println()
	return client.subscriber.Connect(msgQueueURL)
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
