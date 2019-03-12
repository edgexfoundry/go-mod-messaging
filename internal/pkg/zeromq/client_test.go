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
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/pkg/messaging"
	zmq "github.com/pebbe/zmq4"
	"github.com/stretchr/testify/assert"
)

const (
	zeromqPort = 5570
)

var msgConfig = messaging.MessageBusConfig{
	PublishHost: messaging.HostInfo{
		Host:     "*",
		Port:     zeromqPort,
		Protocol: "tcp",
	},
	SubscribeHost: messaging.HostInfo{
		Host:     "localhost",
		Port:     zeromqPort,
		Protocol: "tcp",
	},
}

var zeroMqClient *zeromqClient

func TestMain(m *testing.M) {
	msgConfig.Type = "zero"
	var err error

	zeroMqClient, err = NewZeroMqClient(msgConfig)
	if err != nil {
		fmt.Println("failed to create a new zeromq client")
		os.Exit(-1)
	}
	defer zeroMqClient.Disconnect()
	os.Exit(m.Run())
}

func TestNewClient(t *testing.T) {

	msgConfig.Type = "zero"
	client, err := NewZeroMqClient(msgConfig)

	if err != nil {
		t.Fatal(err)
	}

	if client == nil {
		t.Fatal("Failed to create new zero ZMQ client")
	}

	if client.config.PublishHost.Host != "*" {
		t.Fatal("Failed to populate host value in config")
	}
	if client.config.PublishHost.Port != zeromqPort {
		t.Fatal("Failed to populate port value in config")
	}
	if client.config.PublishHost.Protocol != "tcp" {
		t.Fatal("Failed to populate protocol value in config")
	}

	if client.config.SubscribeHost.Host != "localhost" {
		t.Fatal("Failed to populate host value in config")
	}
	if client.config.SubscribeHost.Port != zeromqPort {
		t.Fatal("Failed to populate port value in config")
	}
	if client.config.SubscribeHost.Protocol != "tcp" {
		t.Fatal("Failed to populate protocol value in config")
	}

}

func TestConnect(t *testing.T) {

	if err := zeroMqClient.Connect(); err != nil {
		t.Fatal("Failed to connect to zero ZMQ")
	}
}

func TestPublish(t *testing.T) {

	zeroMqClient.Connect()

	message := messaging.MessageEnvelope{
		CorrelationID: "123", Payload: []byte("test bytes"),
	}
	topic := ""

	err := zeroMqClient.Publish(message, topic)

	if err != nil {
		t.Fatalf("Failed to publish ZMQ message, %v", err)
	}
}

func TestPublishWithTopic(t *testing.T) {

	zeroMqClient.Connect()

	message := messaging.MessageEnvelope{
		CorrelationID: "123", Payload: []byte("test bytes"),
	}

	topic := "TestTopic"

	err := zeroMqClient.Publish(message, topic)

	if err != nil {
		t.Fatalf("Failed to publish ZMQ message, %v", err)
	}
}

func TestPublishWihEmptyMsg(t *testing.T) {

	zeroMqClient.Connect()

	message := messaging.MessageEnvelope{}

	topic := ""

	err := zeroMqClient.Publish(message, topic)

	if err != nil {
		t.Fatalf("Failed to publish ZMQ message, %v", err)
	}
}

func createAndSubscribeClient(topic string, messages chan *messaging.MessageEnvelope, messageErrors chan error) *zeromqClient {

	testMsgConfig := messaging.MessageBusConfig{
		PublishHost: messaging.HostInfo{
			Host:     "*",
			Port:     5565,
			Protocol: "tcp",
		},
		SubscribeHost: messaging.HostInfo{
			Host:     "localhost",
			Port:     5565,
			Protocol: "tcp",
		},
	}

	client, _ := NewZeroMqClient(testMsgConfig)
	client.Connect()

	topics := []messaging.TopicChannel{{Topic: topic, Messages: messages}}
	zeroMqClient.Subscribe(topics, messageErrors)

	return client
}

func TestCustomPublishWithNoTopic(t *testing.T) {
	zmqClientPort := 5888
	zmqClient, err := getZeroMqClient(zmqClientPort)
	if err != nil {
		t.Fatalf("Failed to create zmqClient with port number %d: %v", zmqClientPort, err)
	}
	defer zmqClient.Disconnect()

	filterTopic := ""
	messages := make(chan *messaging.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []messaging.TopicChannel{{Topic: filterTopic, Messages: messages}}

	err = zmqClient.Subscribe(topics, messageErrors)

	if err != nil {
		t.Fatalf("Failed to subscribe to ZMQ message, %v", err)
	}

	expectedCorreleationID := "123"
	expectedPayload := []byte("test bytes")
	message := messaging.MessageEnvelope{
		CorrelationID: expectedCorreleationID, Payload: expectedPayload,
	}
	dataBytes, err := json.Marshal(message)

	// custom publisher
	customPublisher, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		t.Fatalf("Failed to open publish socket: %v", err)
	}
	publisherMsgQueue := zmqClient.getPublishMessageQueueURL()
	if conErr := customPublisher.Bind(publisherMsgQueue); conErr != nil {
		t.Fatalf("Failed to bind to publisher message queue [%s]: %v", publisherMsgQueue, conErr)
	}

	// publish messages with topic
	time.Sleep(time.Second)
	_, err = customPublisher.SendBytes(dataBytes, 0)
	if err != nil {
		t.Fatalf("Failed to send bytes: %v", err)
	}

	payloadReturned := ""
	testTimer := time.NewTimer(2 * time.Second)
	defer testTimer.Stop()

	done := false
	for !done {
		select {
		case msgErr := <-messageErrors:
			if msgErr == nil {
				done = true
			}
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
		case msgs := <-messages:
			fmt.Printf("In test caller, received messages: %v\n", *msgs)
			payloadReturned = string(msgs.Payload)

			if msgs.CorrelationID != expectedCorreleationID && string(msgs.Payload) == string(expectedPayload) {
				t.Fatal("In test caller, received wrong message")
			}
			done = true
		case <-testTimer.C:
			fmt.Println("timed-out")
			if payloadReturned != "" {
				t.Fatal("Received message with filter on, should have filtered message")
			}
			done = true
		}
	}
	fmt.Println("Done")
}

func TestPublishWihMultipleSubscribers(t *testing.T) {

	topic := ""

	messages1 := make(chan *messaging.MessageEnvelope)
	messageErrors1 := make(chan error)
	client1 := createAndSubscribeClient(topic, messages1, messageErrors1)

	messages2 := make(chan *messaging.MessageEnvelope)
	messageErrors2 := make(chan error)
	_ = createAndSubscribeClient(topic, messages2, messageErrors2)

	// publish messages with topic
	time.Sleep(time.Second)

	expectedCorreleationID := "123"
	expectedPayload := []byte("test bytes")
	message := messaging.MessageEnvelope{
		CorrelationID: expectedCorreleationID, Payload: expectedPayload,
	}

	err := client1.Publish(message, topic)
	if err != nil {
		t.Fatalf("Failed to publish to ZMQ message, %v", err)
	}

	testTimer := time.NewTimer(2 * time.Second)
	defer testTimer.Stop()
	receivedMsg1 := ""
	receivedMsg2 := ""

	for {
		select {
		case msgErr := <-messageErrors1:
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
		case msgs := <-messages1:
			fmt.Printf("Received messages: %v\n", *msgs)
			receivedMsg1 = string(msgs.Payload)
			if msgs.CorrelationID != expectedCorreleationID && string(msgs.Payload) == string(expectedPayload) {
				t.Fatal("Received wrong message")
			}
		case msgErr := <-messageErrors2:
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
		case msgs := <-messages2:
			fmt.Printf("Received messages: %v\n", *msgs)
			receivedMsg2 = string(msgs.Payload)
			if msgs.CorrelationID != expectedCorreleationID && string(msgs.Payload) == string(expectedPayload) {
				t.Fatal("Received wrong message")
			}
		case <-testTimer.C:
			if receivedMsg1 != receivedMsg2 && receivedMsg1 != "" {
				t.Fatal("Received messages don't match")
			}
			return
		}
	}
}

func TestSubscribe(t *testing.T) {

	portNum := 5580

	publishTopic := "testTopic"

	// filter topics
	filterTopics := []string{"", "DONT-MATCH", publishTopic}
	for idx, filterTopic := range filterTopics {
		zmqClient, err := getZeroMqClient(portNum + idx)
		defer zmqClient.Disconnect()
		if err != nil {
			t.Fatalf("Failed to create a new 0mq client with port %d", portNum)
		}
		zmqClient.Connect()
		runPublishSubscribe(t, zmqClient, publishTopic, filterTopic)
	}
}

func getZeroMqClient(zmqPort int) (*zeromqClient, error) {
	zmqConfig := messaging.MessageBusConfig{
		PublishHost: messaging.HostInfo{
			Host:     "127.0.0.1",
			Port:     zmqPort,
			Protocol: "tcp",
		},
		SubscribeHost: messaging.HostInfo{
			Host:     "localhost",
			Port:     zmqPort,
			Protocol: "tcp",
		},
		Type: "zero",
	}

	return NewZeroMqClient(zmqConfig)
}

func runPublishSubscribe(t *testing.T, zmqClient *zeromqClient, publishTopic string, filterTopic string) {

	messages := make(chan *messaging.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []messaging.TopicChannel{{Topic: filterTopic, Messages: messages}}

	err := zmqClient.Subscribe(topics, messageErrors)

	if err != nil {
		t.Fatalf("Failed to subscribe to ZMQ message, %v", err)
	}

	// publish messages with topic
	time.Sleep(time.Second)
	expectedCorreleationID := "123"
	expectedPayload := []byte("test bytes")
	message := messaging.MessageEnvelope{
		CorrelationID: expectedCorreleationID, Payload: expectedPayload,
	}

	err = zmqClient.Publish(message, publishTopic)
	if err != nil {
		t.Fatalf("Failed to publish to ZMQ message, %v", err)
	}

	testTimer := time.NewTimer(3 * time.Second)
	defer testTimer.Stop()
	payloadReturned := ""

	done := false
	for !done {
		select {
		case msgErr := <-messageErrors:
			if msgErr == nil {
				done = true
			}
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
		case msgs := <-messages:
			fmt.Printf("In test caller, received messages: %v\n", *msgs)
			payloadReturned = string(msgs.Payload)

			if msgs.CorrelationID != expectedCorreleationID && string(msgs.Payload) == string(expectedPayload) {
				t.Fatal("In test caller, received wrong message")
			}
			done = true
		case <-testTimer.C:
			fmt.Println("timed-out.")
			if payloadReturned != "" {
				t.Fatal("Received message with filter on, should have filtered message")
			}
			done = true
		}
	}
	fmt.Println("Done")
}

func TestSubscribeMultipleTopics(t *testing.T) {
	zmqClientPort := 5590

	zmqClient, err := getZeroMqClient(zmqClientPort)
	if err != nil {
		t.Fatalf("Failed to create zmqClient with port number %d: %v", zmqClientPort, err)
	}
	defer zmqClient.Disconnect()

	publishTopics := []string{"apple", "orange", "banana", "kiwi"}

	zmqClient.Connect()

	//filterTopics := []string{"test", "DONT", "d"}
	messages1 := make(chan *messaging.MessageEnvelope)
	messages2 := make(chan *messaging.MessageEnvelope)
	messages3 := make(chan *messaging.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []messaging.TopicChannel{
		{Topic: publishTopics[0], Messages: messages1},
		{Topic: publishTopics[1], Messages: messages2},
		{Topic: publishTopics[2], Messages: messages3},
	}

	err = zmqClient.Subscribe(topics, messageErrors)

	if err != nil {
		t.Fatalf("Failed to subscribe to ZMQ message, %v", err)
	}

	// publish messages with topic
	expectedCorreleationID := "123"
	expectedPayload := []byte("test bytes")
	message := messaging.MessageEnvelope{
		CorrelationID: expectedCorreleationID, Payload: expectedPayload,
	}

	time.Sleep(time.Second * 10)
	// publish 4 times:
	for idx := range topics {
		err = zmqClient.Publish(message, publishTopics[idx])
		if err != nil {
			t.Fatalf("Failed to publish to ZMQ message, %v", err)
		}
	}

	testTimer := time.NewTimer(3 * time.Second)
	defer testTimer.Stop()
	payloadReturned := ""

	done := false
	for !done {
		select {
		case msgErr := <-messageErrors:
			if msgErr == nil {
				done = true
			}
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
		case msgs := <-messages1:
			fmt.Printf("In test caller, received messages1: %v\n", *msgs)
			payloadReturned = string(msgs.Payload)

			if msgs.CorrelationID != expectedCorreleationID && string(msgs.Payload) == string(expectedPayload) {
				t.Fatal("In test caller, received wrong message1")
			}
		case msgs := <-messages2:
			fmt.Printf("In test caller, received messages2: %v\n", *msgs)
			payloadReturned = string(msgs.Payload)

			if msgs.CorrelationID != expectedCorreleationID && string(msgs.Payload) == string(expectedPayload) {
				t.Fatal("In test caller, received wrong message2")
			}
		case msgs := <-messages3:
			fmt.Printf("In test caller, received messages3: %v\n", *msgs)
			payloadReturned = string(msgs.Payload)

			if msgs.CorrelationID != expectedCorreleationID && string(msgs.Payload) == string(expectedPayload) {
				t.Fatal("In test caller, received wrong message3")
			}
		case <-testTimer.C:
			fmt.Println("timed-out.")
			if payloadReturned != "" {
				t.Fatal("Received message with filter on, should have filtered message")
			}
			done = true
		}
	}
	fmt.Println("Done")

}

func TestBadSubscriberMessageConfig(t *testing.T) {
	badMsgConfig := messaging.MessageBusConfig{
		SubscribeHost: messaging.HostInfo{
			Host: "\\",
		},
	}

	testClient, err := NewZeroMqClient(badMsgConfig)

	testClient.Connect()
	defer testClient.Disconnect()

	messages := make(chan *messaging.MessageEnvelope)
	topics := []messaging.TopicChannel{{Topic: "", Messages: messages}}
	messageErrors := make(chan error)

	err = testClient.Subscribe(topics, messageErrors)

	if err == nil {
		t.Fatalf("Expecting error from subscribe to ZMQ")
	}
}

func TestBadPublisherMessageConfig(t *testing.T) {
	badMsgConfig := messaging.MessageBusConfig{
		PublishHost: messaging.HostInfo{
			Host: "//",
		},
	}

	testClient, err := NewZeroMqClient(badMsgConfig)

	testClient.Connect()
	defer testClient.Disconnect()

	message := messaging.MessageEnvelope{
		CorrelationID: "123", Payload: []byte("test bytes"),
	}

	topic := "TestTopic"

	err = testClient.Publish(message, topic)

	if err == nil {
		t.Fatalf("Expecting error from publish to ZMQ")
	}
}

func TestDisconnect(t *testing.T) {

	testMsgConfig := messaging.MessageBusConfig{
		PublishHost: messaging.HostInfo{
			Host:     "127.0.0.1",
			Port:     5577,
			Protocol: "tcp",
		},
		SubscribeHost: messaging.HostInfo{
			Host:     "localhost",
			Port:     5577,
			Protocol: "tcp",
		},
	}

	testClient, err := NewZeroMqClient(testMsgConfig)

	testClient.Connect()

	topic := ""
	runPublishSubscribe(t, testClient, topic, "")

	time.Sleep(time.Second * 3)

	err = testClient.Disconnect()

	if assert.NoError(t, err, "Disconnect failed") == false {
		t.Fatal()
	}

	assert.True(t, testClient.publisherDisconnected, "Publisher not closed")
	assert.True(t, testClient.subscriberDisconnected, "Subscriber not closed")
	err = <-testClient.errors
	assert.Nil(t, err, "message error channel is not closed")
	msgEnvelop := <-testClient.topics[0].Messages
	assert.Nil(t, msgEnvelop, "topic channel is not closed")
}

func TestDisconnectError(t *testing.T) {

	testMsgConfig := messaging.MessageBusConfig{
		PublishHost: messaging.HostInfo{
			Host:     "*",
			Port:     5577,
			Protocol: "tcp",
		},
		SubscribeHost: messaging.HostInfo{
			Host:     "localhost",
			Port:     5577,
			Protocol: "tcp",
		},
	}

	testClient, err := NewZeroMqClient(testMsgConfig)

	testClient.Connect()

	topic := ""
	runPublishSubscribe(t, testClient, topic, "")

	time.Sleep(time.Second * 3)

	err = testClient.Disconnect()

	if assert.Error(t, err, "Expecting Disconnect error") == false {
		t.Fatal()
	}
}

func TestGetMsgQueueURL(t *testing.T) {

	url := zeroMqClient.getPublishMessageQueueURL()
	port := strconv.Itoa(zeromqPort)
	if url != "tcp://*:"+port {
		t.Fatal("Failed to create correct publish msg queue URL")
	}

	url = zeroMqClient.getSubscribMessageQueueURL()

	if url != "tcp://localhost:"+port {
		t.Fatal("Failed to create correct subscribe msg queue URL")
	}
}
