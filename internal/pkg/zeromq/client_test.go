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
	"testing"
	"time"

	messaging "github.com/edgexfoundry/go-mod-messaging"
)

var msgConfig = messaging.MessageBusConfig{
	PublishHost: messaging.HostInfo{
		Host:     "*",
		Port:     5563,
		Protocol: "tcp",
	},
	SubscribeHost: messaging.HostInfo{
		Host:     "localhost",
		Port:     5563,
		Protocol: "tcp",
	}}

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
	if client.config.PublishHost.Port != 5563 {
		t.Fatal("Failed to populate port value in config")
	}
	if client.config.PublishHost.Protocol != "tcp" {
		t.Fatal("Failed to populate protocol value in config")
	}

	if client.config.SubscribeHost.Host != "localhost" {
		t.Fatal("Failed to populate host value in config")
	}
	if client.config.SubscribeHost.Port != 5563 {
		t.Fatal("Failed to populate port value in config")
	}
	if client.config.SubscribeHost.Protocol != "tcp" {
		t.Fatal("Failed to populate protocol value in config")
	}

}

func TestConnect(t *testing.T) {

	msgConfig.Type = "zero"
	client, _ := NewZeroMqClient(msgConfig)

	if err := client.Connect(); err != nil {
		t.Fatal("Failed to connect to zero ZMQ")
	}
}

func TestPublish(t *testing.T) {

	msgConfig.Type = "zero"
	client, _ := NewZeroMqClient(msgConfig)

	client.Connect()

	message := messaging.MessageEnvelope{
		CorrelationID: "123", Payload: []byte("test bytes"),
	}
	topic := ""

	err := client.Publish(message, topic)

	if err != nil {
		t.Fatalf("Failed to publish ZMQ message, %v", err)
	}
}

func TestPublishWithTopic(t *testing.T) {

	msgConfig.Type = "zero"
	client, _ := NewZeroMqClient(msgConfig)

	client.Connect()

	message := messaging.MessageEnvelope{
		CorrelationID: "123", Payload: []byte("test bytes"),
	}

	topic := "TestTopic"

	err := client.Publish(message, topic)

	if err != nil {
		t.Fatalf("Failed to publish ZMQ message, %v", err)
	}
}

func TestPublishWihEmptyMsg(t *testing.T) {

	msgConfig.Type = "zero"
	client, _ := NewZeroMqClient(msgConfig)

	client.Connect()

	message := messaging.MessageEnvelope{}

	topic := ""

	err := client.Publish(message, topic)

	if err != nil {
		t.Fatalf("Failed to publish ZMQ message, %v", err)
	}
}

func TestSubscribe(t *testing.T) {

	msgConfig.Type = "zero"
	client, _ := NewZeroMqClient(msgConfig)

	client.Connect()

	messages := make(chan interface{})
	topics := []messaging.TopicChannel{{Topic: "", Messages: messages}}
	messageErrors := make(chan error)

	err := client.Subscribe(topics, "localhost", messageErrors)

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
	err = client.Publish(message, topic)
	if err != nil {
		t.Fatalf("Failed to publish to ZMQ message, %v", err)
	}

	done := false
	for !done {
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
	}
}

func TestGetMsgQueueURL(t *testing.T) {

	url := getMessageQueueURL(&msgConfig.PublishHost)

	if url != "tcp://*:5563" {
		t.Fatal("Failed to create correct msg queue URL")
	}

	url = getMessageQueueURL(&msgConfig.SubscribeHost)

	if url != "tcp://localhost:5563" {
		t.Fatal("Failed to create correct msg queue URL")
	}
}
