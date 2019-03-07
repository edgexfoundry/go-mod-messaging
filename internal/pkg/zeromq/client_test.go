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

	messaging "github.com/edgexfoundry/go-mod-messaging"
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

func TestSubscribe(t *testing.T) {

	zeroMqClient.Connect()

	messages := make(chan interface{})
	topics := []messaging.TopicChannel{{Topic: "", Messages: messages}}
	messageErrors := make(chan error)

	err := zeroMqClient.Subscribe(topics, messageErrors)

	if err != nil {
		t.Fatalf("Failed to subscribe to ZMQ message, %v", err)
	}

	// publish messages with topic
	time.Sleep(time.Second)
	expectedCorreleatedID := "123"
	expectedPayload := []byte("test bytes")
	message := messaging.MessageEnvelope{
		CorrelationID: expectedCorreleatedID, Payload: expectedPayload,
	}

	topic := "TestTopic"
	err = zeroMqClient.Publish(message, topic)
	if err != nil {
		t.Fatalf("Failed to publish to ZMQ message, %v", err)
	}

	done := false
	for !done {
		if messageErrors != nil && messages != nil {
			select {
			case msgErr := <-messageErrors:
				t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
			case msgs := <-messages:
				fmt.Printf("Received messages: %v\n", msgs)
				msgBytes := []byte(msgs.(string))
				var unmarshalledData messaging.MessageEnvelope
				if err := json.Unmarshal(msgBytes, &unmarshalledData); err != nil {
					t.Fatal("Json unmarshal message envelope failed")
				}
				if unmarshalledData.CorrelationID != expectedCorreleatedID && string(unmarshalledData.Payload) == string(expectedPayload) {
					t.Fatal("Received wrong message")
				}
				done = true
			}
		} else {
			break
		}
	}
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

	messages := make(chan interface{})
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
			Host:     "*",
			Port:     5564,
			Protocol: "tcp",
		},
		SubscribeHost: messaging.HostInfo{
			Host:     "localhost",
			Port:     5564,
			Protocol: "tcp",
		},
	}

	testClient, err := NewZeroMqClient(testMsgConfig)

	testClient.Connect()

	messages := make(chan interface{})
	topics := []messaging.TopicChannel{{Topic: "", Messages: messages}}
	messageErrors := make(chan error)

	err = testClient.Subscribe(topics, messageErrors)

	if err != nil {
		t.Fatalf("Failed to subscribe to ZMQ message, %v", err)
	}

	message := messaging.MessageEnvelope{
		CorrelationID: "123", Payload: []byte("test bytes"),
	}
	topic := ""

	err = testClient.Publish(message, topic)

	if assert.NoError(t, err, "Failed to publish ZMQ message") == false {
		t.Fatal()
	}

	done := false
	for !done {
		if messageErrors != nil && messages != nil {
			select {
			case msgErr := <-messageErrors:
				t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
			case msgs := <-messages:
				fmt.Printf("Received messages: %v\n", msgs)
				done = true
			}
		} else {
			break
		}
	}

	err = testClient.Disconnect()

	if assert.NoError(t, err, "Disconnect failed") == false {
		t.Fatal()
	}

	assert.Nil(t, testClient.publisher, "Publisher not closed")
	assert.Nil(t, testClient.subscriber, "Subscriber not closed")

	testerr := <-messageErrors
	assert.NoError(t, testerr, "message error channel is not closed")

	testMessage := <-topics[0].Messages
	assert.Nil(t, testMessage, "topic channel is not closed")
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
