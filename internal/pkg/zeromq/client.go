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
	defaultMsgProtocol       = "tcp"
	maxZeroMqSubscribeTopics = 10
)

type zeromqClient struct {
	publisher             *zmq.Socket
	subscribers           []zeromqSubscriber
	errors                chan error
	config                messaging.MessageBusConfig
	publisherDisconnected bool
}

// NewZeroMqClient instantiates a new zeromq client instance based on the configuration
func NewZeroMqClient(msgConfig messaging.MessageBusConfig) (*zeromqClient, error) {

	client := zeromqClient{config: msgConfig, publisherDisconnected: true}
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

	var mutex = &sync.Mutex{}
	subscriber.waitGrp.Add(1)
	go func(topic *messaging.TopicChannel) {
		defer subscriber.waitGrp.Done()

		mutex.Lock()
		subscriber.connection.SetSubscribe(topic.Topic)
		mutex.Unlock()
		for {
			select {
			case <-subscriber.closed:
				mutex.Lock()
				subscriber.connection.SetUnsubscribe(topic.Topic)
				mutex.Unlock()
				return
			default:
				mutex.Lock()
				payloadMsg, err := subscriber.connection.RecvMessage(zmq4.DONTWAIT)
				mutex.Unlock()

				if err != nil && err.Error() != "resource temporarily unavailable" {
					fmt.Printf("Error received from subscribe: %s", err)
					fmt.Println()
					exiting := false
					if err.Error() == "Socket is closed" {
						exiting = true
					}
					// prevent from blocking if the caller is not reading from the error channel
					go func(quit bool, revErr chan error) {
						revErr <- err
						if quit {
							return
						}
					}(exiting, client.errors)

					if exiting {
						return
					}
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
				//fmt.Printf("receiving message: %v", msgEnvelope)
			}
		}
	}(&subscriber.topic)

	return nil
}

func (client *zeromqClient) areAllSubsribersDisconnected() bool {
	allDisconnected := true
	for _, subscriber := range client.subscribers {
		if !subscriber.disconnected {
			allDisconnected = false
			break
		}
	}
	return allDisconnected
}

func (client *zeromqClient) Disconnect() error {
	// already disconnected, NOP
	if client.publisherDisconnected && client.areAllSubsribersDisconnected() {
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

	for _, subscriber := range client.subscribers {
		if !subscriber.disconnected {
			subscriber.connection.SetLinger(time.Duration(time.Second))

			errSubscribe := subscriber.connection.Disconnect(client.getSubscribMessageQueueURL())
			if errSubscribe != nil {
				fmt.Println("got subscriber disconnect error")
				disconnectErrs = append(disconnectErrs, errSubscribe)
			}

			// close subscriber socket
			errSubscribeClose := subscriber.connection.Close()
			if errSubscribeClose != nil {
				fmt.Println("got subscriber close error")
				disconnectErrs = append(disconnectErrs, errSubscribeClose)
			} else if errSubscribe == nil {
				subscriber.disconnected = true
			}
		}
		subscriber.closed <- true
		subscriber.waitGrp.Wait()
	}

	// close error channel
	if client.errors != nil {
		close(client.errors)
	}

	// close all topic channels
	client.closeTopicChannels()

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

func (client *zeromqClient) closeTopicChannels() {
	for _, subscriber := range client.subscribers {
		if subscriber.topic.Messages != nil {
			close(subscriber.topic.Messages)
		}
		close(subscriber.closed)
	}
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
