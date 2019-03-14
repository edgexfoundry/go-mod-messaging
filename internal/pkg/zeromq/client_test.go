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

func TestMultiplePublishBindsOnSamePortError(t *testing.T) {
	zmqClientPort := 5788
	zmqClient1, err := getZeroMqClient(zmqClientPort)
	if err != nil {
		t.Fatalf("Failed to create zmqClient with port number %d: %v", zmqClientPort, err)
	}
	defer zmqClient1.Disconnect()

	zmqClient2, err := getZeroMqClient(zmqClientPort)
	if err != nil {
		t.Fatalf("Failed to create zmqClient with port number %d: %v", zmqClientPort, err)
	}
	defer zmqClient2.Disconnect()

	message := messaging.MessageEnvelope{
		CorrelationID: "123", Payload: []byte("test bytes"),
	}
	topic := ""

	err = zmqClient1.Publish(message, topic)
	if err != nil {
		t.Fatalf("Failed to publish ZMQ message, %v", err)
	}

	// the second instance of publisher on the same port
	// this should give an error
	err = zmqClient2.Publish(message, topic)
	fmt.Println(err)
	if err == nil {
		t.Fatalf("Expecting to get an error")
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
	testTimer := time.NewTimer(time.Second)
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

			if msgs.CorrelationID != expectedCorreleationID || string(msgs.Payload) != string(expectedPayload) {
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
}

func TestCustomPublishWithWrongMessageLength(t *testing.T) {
	zmqClientPort := 5889
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
	_, err = customPublisher.SendMessage("message1", "message2", "message3")
	if err != nil {
		t.Fatalf("Failed to send multiple messages: %v", err)
	}

	payloadReturned := ""
	testTimer := time.NewTimer(time.Second)
	defer testTimer.Stop()

	done := false
	for !done {
		select {
		case msgErr := <-messageErrors:
			if msgErr == nil {
				t.Fatalf("expecting to get the error for wrong publisher message length")
			}
		case <-messages:
			t.Fatalf("expecting to get the message error for wrong publisher message length and not getting the message here")
		case <-testTimer.C:
			fmt.Println("timed-out")
			if payloadReturned != "" {
				t.Fatal("Received message with filter on, should have filtered message")
			}
			done = true
		}
	}
}

func TestPublishWihMultipleSubscribers(t *testing.T) {

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
	topic := "a"

	messages1 := make(chan *messaging.MessageEnvelope)
	messageErrors1 := make(chan error)
	client1 := createAndSubscribeClient(t, testMsgConfig, topic, messages1, messageErrors1)

	messages2 := make(chan *messaging.MessageEnvelope)
	messageErrors2 := make(chan error)
	_ = createAndSubscribeClient(t, testMsgConfig, topic, messages2, messageErrors2)

	// publish messages with topic
	expectedCorreleationID := "123"
	expectedPayload := []byte("test bytes")
	message := messaging.MessageEnvelope{
		CorrelationID: expectedCorreleationID, Payload: expectedPayload,
	}

	time.Sleep(time.Second * 3)
	err := client1.Publish(message, topic)
	if err != nil {
		t.Fatalf("Failed to publish to ZMQ message, %v", err)
	}

	testTimer := time.NewTimer(3 * time.Second)
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
			if msgs.CorrelationID != expectedCorreleationID || string(msgs.Payload) != string(expectedPayload) {
				t.Fatal("Received wrong message")
			}
		case msgErr := <-messageErrors2:
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
		case msgs := <-messages2:
			fmt.Printf("Received messages: %v\n", *msgs)
			receivedMsg2 = string(msgs.Payload)
			if msgs.CorrelationID != expectedCorreleationID || string(msgs.Payload) != string(expectedPayload) {
				t.Fatal("Received wrong message")
			}
		case <-testTimer.C:
			fmt.Printf("msg1: %s, msg2: %s\n", receivedMsg1, receivedMsg2)

			if receivedMsg1 != receivedMsg2 {
				t.Fatalf("Received messages don't match- msg1 %s  msg2 %s", receivedMsg1, receivedMsg2)
			} else if receivedMsg1 == "" && receivedMsg2 == "" {
				t.Fatal("not received messages")
			}
			return
		}
	}
}

func TestPublishWihMultipleSubscribersWithTopic(t *testing.T) {

	zmqClientPort := 5599

	zmqClient1, err := getZeroMqClient(zmqClientPort)
	if err != nil {
		t.Fatalf("Failed to create zmqClient with port number %d: %v", zmqClientPort, err)
	}

	zmqClient2, err := getZeroMqClient(zmqClientPort)
	if err != nil {
		t.Fatalf("Failed to create zmqClient with port number %d: %v", zmqClientPort, err)
	}

	publishTopics := []string{"apple"}

	zmqClient1.Connect()
	defer zmqClient1.Disconnect()
	zmqClient2.Connect()
	defer zmqClient2.Disconnect()

	messages1 := make(chan *messaging.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []messaging.TopicChannel{
		{Topic: publishTopics[0], Messages: messages1},
	}

	err = zmqClient1.Subscribe(topics, messageErrors)

	if err != nil {
		t.Fatalf("Failed to subscribe to ZMQ message, %v", err)
	}

	// publish messages with topic
	expectedCorreleationIDs := []string{"101"}
	expectedPayloads := [][]byte{[]byte("apple juice")}
	var messages []messaging.MessageEnvelope
	for idx := range expectedCorreleationIDs {
		message := messaging.MessageEnvelope{CorrelationID: expectedCorreleationIDs[idx], Payload: expectedPayloads[idx]}
		messages = append(messages, message)
	}

	time.Sleep(time.Second)
	// publish a few times:
	for idx := range topics {
		err = zmqClient2.Publish(messages[idx], publishTopics[idx])
		if err != nil {
			t.Fatalf("Failed to publish to ZMQ message, %v", err)
		}
	}

	testTimer := time.NewTimer(time.Second)
	defer testTimer.Stop()

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

			if msgs.CorrelationID != expectedCorreleationIDs[0] || string(msgs.Payload) != string(expectedPayloads[0]) {
				t.Fatal("In test caller, received wrong message1")
			}
		case <-testTimer.C:
			fmt.Println("time's up")
			done = true
		}
	}
	fmt.Println("Done")
}

// func TestPublishWihMultipleSubscribersWithTopic(t *testing.T) {

// 	testMsgConfig := messaging.MessageBusConfig{
// 		PublishHost: messaging.HostInfo{
// 			Host:     "*",
// 			Port:     5565,
// 			Protocol: "tcp",
// 		},
// 		SubscribeHost: messaging.HostInfo{
// 			Host:     "localhost",
// 			Port:     5565,
// 			Protocol: "tcp",
// 		},
// 	}

// 	client1Topic := ""

// 	messages1 := make(chan *messaging.MessageEnvelope)
// 	messageErrors1 := make(chan error)
// 	client1 := createAndSubscribeClient(testMsgConfig, client1Topic, messages1, messageErrors1)
// 	defer client1.Disconnect()

// 	client2 := createClient(testMsgConfig)
// 	defer client2.Disconnect()

// 	// publish messages with topic
// 	time.Sleep(time.Second)

// 	expectedCorreleationID := "123"
// 	expectedPayload := []byte("test bytes")
// 	message := messaging.MessageEnvelope{
// 		CorrelationID: expectedCorreleationID, Payload: expectedPayload,
// 	}

// 	// publish to client 1's topic
// 	err := client2.Publish(message, client1Topic)
// 	if err != nil {
// 		t.Fatalf("Failed to publish to ZMQ message, %v", err)
// 	}

// 	testTimer := time.NewTimer(time.Hour)
// 	defer testTimer.Stop()
// 	receivedMsg1 := ""

// 	for {
// 		select {
// 		case msgErr := <-messageErrors1:
// 			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
// 		case msgs := <-messages1:
// 			fmt.Printf("Received messages: %v\n", *msgs)
// 			receivedMsg1 = string(msgs.Payload)
// 			if msgs.CorrelationID != expectedCorreleationID || string(msgs.Payload) != string(expectedPayload) {
// 				t.Fatal("Received wrong message")
// 			}
// 		case <-testTimer.C:
// 			fmt.Printf("msg1: %s\n", receivedMsg1)

// 			if receivedMsg1 == "" {
// 				t.Fatalf("Received message is empty, expecting {%s}\n", string(expectedPayload))
// 			}
// 			return
// 		}
// 	}
// }

func createClient(testMsgConfig messaging.MessageBusConfig) (*zeromqClient, error) {
	client, err := NewZeroMqClient(testMsgConfig)
	return client, err
}

func createAndSubscribeClient(t *testing.T, testMsgConfig messaging.MessageBusConfig, topic string, messages chan *messaging.MessageEnvelope, messageErrors chan error) *zeromqClient {

	client, err := createClient(testMsgConfig)
	if err != nil {
		t.Fatalf("failed to create client with config %v", testMsgConfig)
	}

	client.Connect()

	topics := []messaging.TopicChannel{{Topic: topic, Messages: messages}}
	err = client.Subscribe(topics, messageErrors)
	if err != nil {
		t.Fatalf("failed to subscribe with topic: %v", err)
	}

	return client
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

func TestSubscribeZeroLengthTopic(t *testing.T) {
	portNum := 5780

	zmqClient, err := getZeroMqClient(portNum)
	if err != nil {
		t.Fatalf("Failed to create a new 0mq client with port %d", portNum)
	}
	zmqClient.Connect()
	defer zmqClient.Disconnect()

	err = zmqClient.Subscribe(nil, make(chan error))

	if err == nil {
		t.Fatalf("Expecting to get error on nil topics to subscribe to ZMQ message")
	}
}

func TestSubscribeExceedsMaxNumberTopic(t *testing.T) {
	portNum := 5781

	zmqClient, err := getZeroMqClient(portNum)
	if err != nil {
		t.Fatalf("Failed to create a new 0mq client with port %d", portNum)
	}
	zmqClient.Connect()
	defer zmqClient.Disconnect()

	mockTopics := make([]messaging.TopicChannel, maxZeroMqSubscribeTopics+1)
	err = zmqClient.Subscribe(mockTopics, make(chan error))

	if err == nil {
		t.Fatalf("Expecting to get error on exceeding max number of topics to subscribe to ZMQ message")
	}
}

func getZeroMqClient(zmqPort int) (*zeromqClient, error) {
	zmqConfig := messaging.MessageBusConfig{
		PublishHost: messaging.HostInfo{
			Host:     "*",
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
	exptectedContentType := "application/json"
	message := messaging.MessageEnvelope{
		CorrelationID: expectedCorreleationID, Payload: expectedPayload,
		ContentType: exptectedContentType,
	}

	err = zmqClient.Publish(message, publishTopic)
	if err != nil {
		t.Fatalf("Failed to publish to ZMQ message, %v", err)
	}

	testTimer := time.NewTimer(time.Second)
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

			if msgs.CorrelationID != expectedCorreleationID ||
				string(msgs.Payload) != string(expectedPayload) ||
				msgs.ContentType != exptectedContentType {
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
}

func TestSubscribeMultipleTopics(t *testing.T) {

	zmqClientPort := 5590

	zmqClient, err := getZeroMqClient(zmqClientPort)
	if err != nil {
		t.Fatalf("Failed to create zmqClient with port number %d: %v", zmqClientPort, err)
	}

	publishTopics := []string{"apple", "orange", "banana"}

	zmqClient.Connect()
	defer zmqClient.Disconnect()

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
	expectedCorreleationIDs := []string{"101", "102", "103"}
	expectedPayloads := [][]byte{[]byte("apple juice"), []byte("orange juice"), []byte("banana slices")}
	exptectedContentTypes := []string{"application/json", "application/cbor", "video"}
	var messages []messaging.MessageEnvelope
	for idx := range expectedCorreleationIDs {
		message := messaging.MessageEnvelope{
			CorrelationID: expectedCorreleationIDs[idx],
			Payload:       expectedPayloads[idx],
			ContentType:   exptectedContentTypes[idx]}
		messages = append(messages, message)
	}

	time.Sleep(time.Second)
	// publish a few times:
	for idx := range topics {
		err = zmqClient.Publish(messages[idx], publishTopics[idx])
		if err != nil {
			t.Fatalf("Failed to publish to ZMQ message, %v", err)
		}
	}

	testTimer := time.NewTimer(time.Second)
	defer testTimer.Stop()

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

			if msgs.CorrelationID != expectedCorreleationIDs[0] ||
				string(msgs.Payload) != string(expectedPayloads[0]) ||
				msgs.ContentType != exptectedContentTypes[0] {
				t.Fatal("In test caller, received wrong message1")
			}
		case msgs := <-messages2:
			fmt.Printf("In test caller, received messages2: %v\n", *msgs)

			if msgs.CorrelationID != expectedCorreleationIDs[1] ||
				string(msgs.Payload) != string(expectedPayloads[1]) ||
				msgs.ContentType != exptectedContentTypes[1] {
				t.Fatal("In test caller, received wrong message2")
			}
		case msgs := <-messages3:
			fmt.Printf("In test caller, received messages3: %v\n", *msgs)

			if msgs.CorrelationID != expectedCorreleationIDs[2] ||
				string(msgs.Payload) != string(expectedPayloads[2]) ||
				msgs.ContentType != exptectedContentTypes[2] {
				t.Fatal("In test caller, received wrong message3")
			}
		case <-testTimer.C:
			fmt.Println("time's up")
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

	messages := make(chan *messaging.MessageEnvelope)
	topics := []messaging.TopicChannel{{Topic: "", Messages: messages}}
	messageErrors := make(chan error)

	err = testClient.Subscribe(topics, messageErrors)

	if err == nil {
		defer testClient.Disconnect()
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

	err = testClient.Disconnect()

	if assert.NoError(t, err, "Disconnect failed") == false {
		t.Fatal()
	}

	assert.True(t, testClient.allSubscribersDisconnected(), "Subscriber not closed")

	err = <-testClient.errors
	assert.Nil(t, err, "message error channel is not closed")
	msgEnvelop := <-testClient.subscribers[0].topic.Messages
	assert.Nil(t, msgEnvelop, "topic channel is not closed")
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
