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

	"github.com/edgexfoundry/go-mod-messaging/pkg/messaging"

	"github.com/pebbe/zmq4"
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
	publisher              *zmq.Socket
	subscriber             *zmq.Socket
	topics                 []messaging.TopicChannel
	errors                 chan error
	config                 messaging.MessageBusConfig
	closed                 chan bool
	waitGrp                sync.WaitGroup
	publisherDisconnected  bool
	subscriberDisconnected bool
}

// NewZeroMqClient instantiates a new zeromq client instance based on the configuration
func NewZeroMqClient(msgConfig messaging.MessageBusConfig) (*zeromqClient, error) {

	client := zeromqClient{config: msgConfig, closed: make(chan bool), publisherDisconnected: true, subscriberDisconnected: true}
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
			return conErr
		}

		client.publisherDisconnected = false
		fmt.Println("Publisher successfully connected to 0MQ message queue")

		// allow some time for socket binding before start publishing
		time.Sleep(time.Second)
	} else if client.publisherDisconnected {
		// reconnect with the endpoint
		if conErr := client.publisher.Bind(msgQueueURL); conErr != nil {
			return conErr
		}
		client.publisherDisconnected = false
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

	client.subscriberDisconnected = false

	for _, topic := range topics {
		client.subscriber.SetSubscribe(topic.Topic)

		client.waitGrp.Add(1)
		go func(topic messaging.TopicChannel) {
			defer client.waitGrp.Done()

			for {
				select {
				case <-client.closed:
					return
				default:
					payloadMsg, err := client.subscriber.RecvMessage(zmq4.DONTWAIT)

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
					fmt.Printf("receiving message: %v", msgEnvelope)
				}
			}
		}(topic)
	}

	return nil
}

func (client *zeromqClient) Disconnect() error {
	// already disconnected, NOP
	if client.publisherDisconnected && client.subscriberDisconnected {
		return nil
	}

	var disconnectErrs []error
	// use 0mq's disconnect / unbind API to disconnect from the endpoint
	if client.publisher != nil && !client.publisherDisconnected {
		client.publisher.SetLinger(time.Duration(time.Second))

		errPublish := client.publisher.Unbind(client.getPublishMessageQueueURL())
		if errPublish != nil {
			fmt.Println("got publisher disconnect error")
			disconnectErrs = append(disconnectErrs, errPublish)
		}

		// close publisher socket
		errPublishClose := client.publisher.Close()
		if errPublishClose != nil {
			fmt.Println("got publisher close error")
			disconnectErrs = append(disconnectErrs, errPublishClose)
		} else if errPublish == nil {
			client.publisherDisconnected = true
		}
	}

	if client.subscriber != nil && !client.subscriberDisconnected {
		client.subscriber.SetLinger(time.Duration(time.Second))

		errSubscribe := client.subscriber.Disconnect(client.getSubscribMessageQueueURL())
		if errSubscribe != nil {
			fmt.Println("got subscriber disconnect error")
			disconnectErrs = append(disconnectErrs, errSubscribe)
		}

		// close subscriber socket
		errSubscribeClose := client.subscriber.Close()
		if errSubscribeClose != nil {
			fmt.Println("got subscriber close error")
			disconnectErrs = append(disconnectErrs, errSubscribeClose)
		} else if errSubscribe == nil {
			client.subscriberDisconnected = true
		}
	}

	client.closed <- true
	client.waitGrp.Wait()

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

	close(client.closed)

	// waiting for some time before we return for cleaning up takes time
	time.Sleep(time.Second * 2)

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
