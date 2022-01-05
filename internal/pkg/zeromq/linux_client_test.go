//go:build linux
// +build linux
//
// Copyright (c) 2021 Intel Corporation
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
	"testing"
	"time"

	zmq "github.com/pebbe/zmq4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

const (
	zeromqPort = 5570
)

var msgConfig = types.MessageBusConfig{
	PublishHost: types.HostInfo{
		Host:     "*",
		Port:     zeromqPort,
		Protocol: "tcp",
	},
	SubscribeHost: types.HostInfo{
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
		fmt.Println("Failed to create a new zeromq client")
		os.Exit(-1)
	}
	defer func() { _ = zeroMqClient.Disconnect() }()
	os.Exit(m.Run())
}

func TestNewClient(t *testing.T) {

	msgConfig.Type = "zero"
	client, err := NewZeroMqClient(msgConfig)

	if !assert.Nil(t, err, "Error creating to create ZMQ client") {
		t.Fatal()
	}
	if !assert.NotNil(t, client, "Failed to create ZMQ client") {
		t.Fatal()
	}
	if !assert.Equal(t, "*", client.config.PublishHost.Host, "Failed to populate host value in config") {
		t.Fatal()
	}
	if !assert.Equal(t, zeromqPort, client.config.PublishHost.Port, "Failed to populate port value in config") {
		t.Fatal()
	}
	if !assert.Equal(t, "tcp", client.config.PublishHost.Protocol, "Failed to populate protocol value in config") {
		t.Fatal()
	}
	if !assert.Equal(t, "localhost", client.config.SubscribeHost.Host, "Failed to populate host value in config") {
		t.Fatal()
	}
	if !assert.Equal(t, zeromqPort, client.config.SubscribeHost.Port, "Failed to populate port value in config") {
		t.Fatal()
	}
	if !assert.Equal(t, "tcp", client.config.SubscribeHost.Protocol, "Failed to populate protocol value in config") {
		t.Fatal()
	}
}

func TestConnect(t *testing.T) {

	err := zeroMqClient.Connect()
	if !assert.Nil(t, err, "Failed to connect to zero ZMQ") {
		t.Fatal()
	}
}

func TestPublish(t *testing.T) {

	_ = zeroMqClient.Connect()

	message := types.MessageEnvelope{
		CorrelationID: "123", Payload: []byte("test bytes"),
	}
	topic := ""

	err := zeroMqClient.Publish(message, topic)

	if !assert.Nil(t, err, "Failed to publish ZMQ message") {
		t.Fatal()
	}
}

func TestMultiplePublishBindsOnSamePortError(t *testing.T) {
	zmqClientPort := 5788
	zmqClient1, err := getZeroMqClient(zmqClientPort)

	if !assert.Nil(t, err, "Failed to create zmqClient1") {
		t.Fatal()
	}

	defer func() { _ = zmqClient1.Disconnect() }()

	zmqClient2, err := getZeroMqClient(zmqClientPort)

	if !assert.Nil(t, err, "Failed to create zmqClient2") {
		t.Fatal()
	}

	defer func() { _ = zmqClient2.Disconnect() }()

	message := types.MessageEnvelope{
		CorrelationID: "123", Payload: []byte("test bytes"),
	}
	topic := ""

	err = zmqClient1.Publish(message, topic)

	if !assert.Nil(t, err, "Failed to publish ZMQ message") {
		t.Fatal()
	}

	// the second instance of publisher on the same port
	// this should give an error
	err = zmqClient2.Publish(message, topic)

	if !assert.NotNil(t, err, "Expecting to get an error publishing") {
		t.Fatal()
	}
}

func TestPublishWithTopic(t *testing.T) {

	_ = zeroMqClient.Connect()

	message := types.MessageEnvelope{
		CorrelationID: "123", Payload: []byte("test bytes"),
	}

	topic := "TestTopic"

	err := zeroMqClient.Publish(message, topic)

	if !assert.Nil(t, err, "Failed to publish ZMQ message") {
		t.Fatal()
	}
}

func TestPublishWihEmptyMsg(t *testing.T) {

	_ = zeroMqClient.Connect()

	message := types.MessageEnvelope{}

	topic := ""

	err := zeroMqClient.Publish(message, topic)

	if !assert.Nil(t, err, "Failed to publish ZMQ message") {
		t.Fatal()
	}
}

func TestCustomPublishWithNoTopic(t *testing.T) {
	zmqClientPort := 5888
	zmqClient, err := getZeroMqClient(zmqClientPort)

	if !assert.Nil(t, err, "Failed to create zmqClient") {
		t.Fatal()
	}

	defer func() { _ = zmqClient.Disconnect() }()

	filterTopic := "filter"
	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []types.TopicChannel{{Topic: filterTopic, Messages: messages}}

	err = zmqClient.Subscribe(topics, messageErrors)

	if !assert.Nil(t, err, "Failed to subscribe to ZMQ message") {
		t.Fatal()
	}

	expectedCorreleationID := "123"
	expectedPayload := []byte("test bytes")
	message := types.MessageEnvelope{
		CorrelationID: expectedCorreleationID, Payload: expectedPayload,
	}
	dataBytes, err := json.Marshal(message)
	require.NoError(t, err)

	// custom publisher
	customPublisher, err := zmq.NewSocket(zmq.PUB)

	if !assert.Nil(t, err, "Failed to open publish socket") {
		t.Fatal()
	}

	publisherMsgQueue := zmqClient.config.PublishHost.GetHostURL()

	conErr := customPublisher.Bind(publisherMsgQueue)

	if !assert.Nilf(t, conErr, "Failed to bind to publisher message queue [%s]", publisherMsgQueue) {
		t.Fatal()
	}

	// publish messages with topic
	time.Sleep(time.Second)
	_, err = customPublisher.SendBytes(dataBytes, 0)

	if !assert.Nil(t, err, "Failed to send bytes") {
		t.Fatal()
	}

	payloadReturned := ""
	testTimer := time.NewTimer(time.Second)
	defer func() { _ = testTimer.Stop() }()

	done := false
	for !done {
		select {
		case msgErr := <-messageErrors:
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)

		case msgs := <-messages:
			fmt.Printf("In test caller, received messages: %v\n", msgs)
			payloadReturned = string(msgs.Payload)

			require.Equal(t, expectedCorreleationID, msgs.CorrelationID)
			require.Equal(t, string(expectedPayload), string(msgs.Payload))
			require.Equal(t, filterTopic, msgs.ReceivedTopic)
			done = true

		case <-testTimer.C:
			fmt.Println("timed-out")

			if !assert.Empty(t, payloadReturned) {
				t.Fatal("Received message with filter on, should have filtered message")
			}
			done = true
		}
	}
}

func TestCustomPublishWithWrongMessageLength(t *testing.T) {
	zmqClientPort := 5889
	zmqClient, err := getZeroMqClient(zmqClientPort)

	if !assert.Nil(t, err, "Failed to create zmqClient") {
		t.Fatal()
	}

	defer func() { _ = zmqClient.Disconnect() }()

	filterTopic := ""
	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []types.TopicChannel{{Topic: filterTopic, Messages: messages}}

	err = zmqClient.Subscribe(topics, messageErrors)

	if !assert.Nil(t, err, "Failed to subscribe to ZMQ") {
		t.Fatal()
	}

	// custom publisher
	customPublisher, err := zmq.NewSocket(zmq.PUB)

	if !assert.Nil(t, err, "Failed to open publish socket") {
		t.Fatal()
	}

	publisherMsgQueue := zmqClient.config.PublishHost.GetHostURL()

	conErr := customPublisher.Bind(publisherMsgQueue)

	if !assert.Nil(t, conErr, "Failed to bind to publisher message queue") {
		t.Fatal()
	}

	// publish messages with topic
	time.Sleep(time.Second)

	_, err = customPublisher.SendMessage("message1", "message2", "message3")

	if !assert.Nil(t, err, "Failed to send multiple messages") {
		t.Fatal()
	}

	payloadReturned := ""
	testTimer := time.NewTimer(time.Second)
	defer func() { _ = testTimer.Stop() }()

	done := false
	for !done {
		select {
		case msgErr := <-messageErrors:
			if !assert.NotNil(t, msgErr) {
				t.Fatal("Expecting to get the error for wrong publisher message length")
			}
		case <-messages:
			t.Fatalf("expecting to get the message error for wrong publisher message length and not getting the message here")
		case <-testTimer.C:
			fmt.Println("timed-out")

			if !assert.Empty(t, payloadReturned) {
				t.Fatal("Received message with filter on, should have filtered message")
			}
			done = true
		}
	}
}

func TestPublishWihMultipleSubscribers(t *testing.T) {

	clientPort := 5585
	publishTopic := ""

	client1, err := getZeroMqClient(clientPort)

	if !assert.Nil(t, err, "Failed to create client1") {
		t.Fatal()
	}

	client2, err := getZeroMqClient(clientPort)

	if !assert.Nil(t, err, "Failed to create client2") {
		t.Fatal()
	}

	_ = client1.Connect()
	defer func() { _ = client1.Disconnect() }()
	_ = client2.Connect()
	defer func() { _ = client2.Disconnect() }()

	messages1 := make(chan types.MessageEnvelope)
	messageErrors1 := make(chan error)
	topics1 := []types.TopicChannel{
		{Topic: publishTopic, Messages: messages1},
	}

	messages2 := make(chan types.MessageEnvelope)
	messageErrors2 := make(chan error)
	topics2 := []types.TopicChannel{
		{Topic: publishTopic, Messages: messages2},
	}

	err = client1.Subscribe(topics1, messageErrors1)

	if !assert.Nil(t, err, "Failed to subscribe to ZMQ message for client1") {
		t.Fatal()
	}

	err = client2.Subscribe(topics2, messageErrors2)

	if !assert.Nil(t, err, "Failed to subscribe to ZMQ message for client1") {
		t.Fatal()
	}

	// publish messages with topic
	expectedCorreleationID := "123"
	expectedPayload := []byte("test bytes")
	message := types.MessageEnvelope{
		CorrelationID: expectedCorreleationID, Payload: expectedPayload,
	}

	err = client1.Publish(message, publishTopic)

	if !assert.Nil(t, err, "Failed to publish to ZMQ message") {
		t.Fatal()
	}

	testTimer := time.NewTimer(3 * time.Second)
	defer func() { _ = testTimer.Stop() }()
	receivedMsg1 := ""
	receivedMsg2 := ""

	for {
		select {
		case msgErr := <-messageErrors1:
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
		case msgs := <-messages1:
			fmt.Printf("Received messages: %v\n", msgs)
			receivedMsg1 = string(msgs.Payload)
			require.Equal(t, expectedCorreleationID, msgs.CorrelationID, "CorreleationIDs don't match")
			require.Equal(t, string(expectedPayload), string(msgs.Payload), "Payloads don't match")
			require.Equal(t, publishTopic, msgs.ReceivedTopic)

		case msgErr := <-messageErrors2:
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)

		case msgs := <-messages2:
			fmt.Printf("Received messages: %v\n", msgs)
			receivedMsg2 = string(msgs.Payload)
			require.Equal(t, expectedCorreleationID, msgs.CorrelationID, "CorreleationIDs don't match")
			require.Equal(t, string(expectedPayload), string(msgs.Payload), "Payloads don't match")
			require.Equal(t, publishTopic, msgs.ReceivedTopic)

		case <-testTimer.C:
			fmt.Printf("msg1: %s, msg2: %s\n", receivedMsg1, receivedMsg2)

			if !assert.Equalf(t, receivedMsg1, receivedMsg2, "Received messages don't match- msg1 %s  msg2 %s", receivedMsg1, receivedMsg2) {
				t.Fatal()
			}
			if !assert.NotEmpty(t, receivedMsg1, "Received messages are empty") {
				t.Fatal()
			}
			return
		}
	}
}

func TestPublishWihMultipleSubscribersWithTopic(t *testing.T) {

	zmqClientPort1 := 5612
	zmqClientPort2 := 5533

	zmqClientCoreData, err := getZeroMqClient(zmqClientPort1)
	if !assert.Nil(t, err, "Failed to create zmqClientCoreData") {
		t.Fatal()
	}

	zmqClientAppFunc1, err := getZeroMqClient(zmqClientPort2)
	if !assert.Nil(t, err, "Failed to create zmqClientAppFunc1") {
		t.Fatal()
	}

	zmqClientAppFunc2, err := getZeroMqClient(zmqClientPort2)
	if !assert.Nil(t, err, "Failed to create zmqClientAppFunc2") {
		t.Fatal()
	}

	_ = zmqClientCoreData.Connect()
	defer func() { _ = zmqClientCoreData.Disconnect() }()
	_ = zmqClientAppFunc1.Connect()
	defer func() { _ = zmqClientAppFunc1.Disconnect() }()
	_ = zmqClientAppFunc2.Connect()
	defer func() { _ = zmqClientAppFunc2.Disconnect() }()

	coreDataPublishTopic := "orange"
	appFunctionPublishTopic := "apple"
	coreDataMessage := types.MessageEnvelope{CorrelationID: "123", Payload: []byte("orange juice")}
	appFuncMessage := types.MessageEnvelope{CorrelationID: "456", Payload: []byte("apple juice")}

	appFuncMessages1 := make(chan types.MessageEnvelope)
	appFuncMessageErrors1 := make(chan error)
	appFuncTopic1 := []types.TopicChannel{
		{Topic: coreDataPublishTopic, Messages: appFuncMessages1},
	}

	appFuncMessages2 := make(chan types.MessageEnvelope)
	appFuncMessageErrors2 := make(chan error)
	appFuncTopic2 := []types.TopicChannel{
		{Topic: appFunctionPublishTopic, Messages: appFuncMessages2},
	}

	err = zmqClientAppFunc1.Subscribe(appFuncTopic1, appFuncMessageErrors1)
	require.NoError(t, err, "Failed to subscribe to ZMQ message")

	err = zmqClientAppFunc2.Subscribe(appFuncTopic2, appFuncMessageErrors2)
	require.NoError(t, err, "Failed to subscribe to ZMQ message")

	time.Sleep(time.Second)

	err = zmqClientCoreData.Publish(coreDataMessage, coreDataPublishTopic)
	require.NoError(t, err, "Failed to publish ZMQ message")

	err = zmqClientAppFunc1.Publish(appFuncMessage, appFunctionPublishTopic)
	require.NoError(t, err, "Failed to publish ZMQ message")

	testTimer := time.NewTimer(time.Second)
	defer func() { _ = testTimer.Stop() }()

	done := false
	for !done {
		select {
		case msgErr := <-appFuncMessageErrors1:
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
		case msgs := <-appFuncMessages1:
			fmt.Printf("App Functions 1 Message: %v\n", msgs)
			require.Equal(t, appFuncMessage.CorrelationID, msgs.CorrelationID)
			require.Equal(t, string(appFuncMessage.Payload), string(msgs.Payload))
			require.Equal(t, coreDataPublishTopic, msgs.ReceivedTopic)

		case msgErr := <-appFuncMessageErrors2:
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
		case msgs := <-appFuncMessages2:
			fmt.Printf("App Functions 2 Message: %v\n", msgs)
			require.Equal(t, appFuncMessage.CorrelationID, msgs.CorrelationID)
			require.Equal(t, string(appFuncMessage.Payload), string(msgs.Payload))
			require.Equal(t, appFunctionPublishTopic, msgs.ReceivedTopic)

		case <-testTimer.C:
			fmt.Println("time's up")
			done = true
		}
	}
	fmt.Println("Done")
}

func TestSubscribe(t *testing.T) {

	portNum := 5580

	publishTopic := "testTopic"

	// filter topics
	filterTopics := []string{"", "DONT-MATCH", publishTopic}
	for idx, filterTopic := range filterTopics {
		zmqClient, err := getZeroMqClient(portNum + idx)
		if !assert.Nil(t, err, "Failed to create a new ZMQ client") {
			t.Fatal()
		}
		defer func() { _ = zmqClient.Disconnect() }()

		_ = zmqClient.Connect()
		runPublishSubscribe(t, zmqClient, publishTopic, filterTopic)
	}
}

func TestSubscribeZeroLengthTopic(t *testing.T) {
	portNum := 5780

	zmqClient, err := getZeroMqClient(portNum)
	if !assert.Nil(t, err, "Failed to create a new ZMQ client") {
		t.Fatal()
	}

	_ = zmqClient.Connect()
	defer func() { _ = zmqClient.Disconnect() }()

	err = zmqClient.Subscribe(nil, make(chan error))

	if !assert.NotNil(t, err, "Expecting to get error when subscribing to nil topics") {
		t.Fatal()
	}
}

func TestSubscribeExceedsMaxNumberTopic(t *testing.T) {
	portNum := 5781

	zmqClient, err := getZeroMqClient(portNum)
	if !assert.Nil(t, err, "Failed to create a new ZMQ client") {
		t.Fatal()
	}

	_ = zmqClient.Connect()
	defer func() { _ = zmqClient.Disconnect() }()

	mockTopics := make([]types.TopicChannel, maxZeroMqSubscribeTopics+1)
	err = zmqClient.Subscribe(mockTopics, make(chan error))

	assert.NotNil(t, err, "Expecting to get error on exceeding max number of topics to subscribe to ZMQ message")
}

func getZeroMqClient(zmqPort int) (*zeromqClient, error) {
	zmqConfig := types.MessageBusConfig{
		PublishHost: types.HostInfo{
			Host:     "*",
			Port:     zmqPort,
			Protocol: "tcp",
		},
		SubscribeHost: types.HostInfo{
			Host:     "localhost",
			Port:     zmqPort,
			Protocol: "tcp",
		},
		Type: "zero",
	}

	return NewZeroMqClient(zmqConfig)
}

func runPublishSubscribe(t *testing.T, zmqClient *zeromqClient, publishTopic string, filterTopic string) {

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []types.TopicChannel{{Topic: filterTopic, Messages: messages}}

	err := zmqClient.Subscribe(topics, messageErrors)

	if !assert.Nil(t, err, "Failed to subscribe to ZMQ message") {
		t.Fatal()
	}

	// publish messages with topic
	time.Sleep(time.Second)
	expectedCorreleationID := "123"
	expectedPayload := []byte("test bytes")
	expectedContentType := "application/json"
	message := types.MessageEnvelope{
		CorrelationID: expectedCorreleationID, Payload: expectedPayload,
		ContentType: expectedContentType,
	}

	err = zmqClient.Publish(message, publishTopic)
	if !assert.Nil(t, err, "Failed to publish to ZMQ message") {
		t.Fatal()
	}

	testTimer := time.NewTimer(time.Second)
	defer func() { _ = testTimer.Stop() }()
	payloadReturned := ""

	done := false
	for !done {
		select {
		case msgErr := <-messageErrors:
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
		case msgs := <-messages:
			fmt.Printf("In test caller, received messages: %v\n", msgs)
			payloadReturned = string(msgs.Payload)

			if !assert.Equal(t, expectedCorreleationID, msgs.CorrelationID) ||
				!assert.Equal(t, string(expectedPayload), string(msgs.Payload)) {
				t.Fatal("In test caller, received wrong message")
			}
			done = true
		case <-testTimer.C:
			fmt.Println("timed-out.")

			if !assert.Empty(t, payloadReturned, "Failed to receive message") {
				t.Fatal()
			}
			done = true
		}
	}
}

func TestSubscribeMultipleTopics(t *testing.T) {

	zmqClientPort := 5590

	zmqClient, err := getZeroMqClient(zmqClientPort)
	if !assert.Nil(t, err, "Failed to create new zmqClient") {
		t.Fatal()
	}

	publishTopics := []string{"apple", "orange", "banana"}

	_ = zmqClient.Connect()
	defer func() { _ = zmqClient.Disconnect() }()

	messages1 := make(chan types.MessageEnvelope)
	messages2 := make(chan types.MessageEnvelope)
	messages3 := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []types.TopicChannel{
		{Topic: publishTopics[0], Messages: messages1},
		{Topic: publishTopics[1], Messages: messages2},
		{Topic: publishTopics[2], Messages: messages3},
	}

	err = zmqClient.Subscribe(topics, messageErrors)
	if !assert.Nil(t, err, "Failed to subscribe to ZMQ message") {
		t.Fatal()
	}

	// publish messages with topic
	expectedCorreleationIDs := []string{"101", "102", "103"}
	expectedPayloads := [][]byte{[]byte("apple juice"), []byte("orange juice"), []byte("banana slices")}
	expectedContentTypes := []string{"application/json", "application/cbor", "video"}
	var messages []types.MessageEnvelope
	for idx := range expectedCorreleationIDs {
		message := types.MessageEnvelope{
			CorrelationID: expectedCorreleationIDs[idx],
			Payload:       expectedPayloads[idx],
			ContentType:   expectedContentTypes[idx]}
		messages = append(messages, message)
	}

	time.Sleep(time.Second)
	// publish a few times:
	for idx := range topics {
		err = zmqClient.Publish(messages[idx], publishTopics[idx])
		if !assert.Nil(t, err, "Failed to publish to ZMQ message") {
			t.Fatal()
		}
	}

	testTimer := time.NewTimer(time.Second)
	defer func() { _ = testTimer.Stop() }()

	done := false
	for !done {
		select {
		case msgErr := <-messageErrors:
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)

		case msgs := <-messages1:
			fmt.Printf("In test caller, received messages1: %v\n", msgs)
			require.Equal(t, expectedCorreleationIDs[0], msgs.CorrelationID)
			require.Equal(t, string(expectedPayloads[0]), string(msgs.Payload))
			require.Equal(t, publishTopics[0], msgs.ReceivedTopic)

		case msgs := <-messages2:
			fmt.Printf("In test caller, received messages2: %v\n", msgs)
			require.Equal(t, expectedCorreleationIDs[1], msgs.CorrelationID)
			require.Equal(t, string(expectedPayloads[1]), string(msgs.Payload))
			require.Equal(t, publishTopics[1], msgs.ReceivedTopic)

		case msgs := <-messages3:
			fmt.Printf("In test caller, received messages3: %v\n", msgs)
			require.Equal(t, expectedCorreleationIDs[2], msgs.CorrelationID)
			require.Equal(t, string(expectedPayloads[2]), string(msgs.Payload))
			require.Equal(t, publishTopics[2], msgs.ReceivedTopic)

		case <-testTimer.C:
			fmt.Println("time's up")
			done = true
		}
	}
	fmt.Println("Done")
}

func TestSubscribeMultipleAndEmptyTopic(t *testing.T) {

	zmqPort := 5599

	publishClient, err := getZeroMqClient(zmqPort)
	if !assert.Nil(t, err, "Failed to create publishClient") {
		t.Fatal()
	}

	_ = publishClient.Connect()
	defer func() { _ = publishClient.Disconnect() }()

	subscribeClient1, err := getZeroMqClient(zmqPort)
	if !assert.Nil(t, err, "Failed to create subscribeClient1") {
		t.Fatal()
	}

	_ = subscribeClient1.Connect()
	defer func() { _ = subscribeClient1.Disconnect() }()

	subscribeClient2, err := getZeroMqClient(zmqPort)
	if !assert.Nil(t, err, "Failed to create subscribeClient2") {
		t.Fatal()
	}

	_ = subscribeClient2.Connect()
	defer func() { _ = subscribeClient2.Disconnect() }()

	publishTopic := "goldfish"
	subscribeTopic1 := publishTopic
	subscribeTopic2 := ""

	messages1 := make(chan types.MessageEnvelope)
	messages2 := make(chan types.MessageEnvelope)
	messageErrors1 := make(chan error)
	messageErrors2 := make(chan error)

	err = subscribeClient1.Subscribe([]types.TopicChannel{{Topic: subscribeTopic1, Messages: messages1}}, messageErrors1)
	if !assert.Nil(t, err, "Failed to subscribe to ZMQ message") {
		t.Fatal()
	}

	err = subscribeClient2.Subscribe([]types.TopicChannel{{Topic: subscribeTopic2, Messages: messages2}}, messageErrors2)
	if !assert.Nil(t, err, "Failed to subscribe to ZMQ message") {
		t.Fatal()
	}

	message1 := types.MessageEnvelope{
		CorrelationID: "123",
		Payload:       []byte("yellow goldfish"),
		ContentType:   "application/json"}
	message2 := types.MessageEnvelope{
		CorrelationID: "123",
		Payload:       []byte("black guppy"),
		ContentType:   "application/json"}

	time.Sleep(time.Second)
	// publish both messages:
	err = publishClient.Publish(message1, publishTopic)
	if !assert.Nil(t, err, "Failed to publish messages to ZMQ") {
		t.Fatal()
	}

	err = publishClient.Publish(message2, "")
	if !assert.Nil(t, err, "Failed to publish message to ZMQ") {
		t.Fatal()
	}

	testTimer := time.NewTimer(time.Second)
	defer func() { _ = testTimer.Stop() }()

	done := false
	receivedMsgs := 0
	expectingMsgs := 3
	for !done {
		select {
		case msgErr := <-messageErrors1:
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
		case msgErr := <-messageErrors2:
			t.Fatalf("Failed to receive ZMQ message, %v", msgErr)
		case msgs := <-messages1:
			fmt.Printf("In test caller, received messages1: %v\n", string(msgs.Payload))
			receivedMsgs++
		case msgs := <-messages2:
			fmt.Printf("In test caller, received messages2: %v\n", string(msgs.Payload))
			receivedMsgs++
		case <-testTimer.C:
			fmt.Println("time's up")
			if !assert.Equalf(t, 3, receivedMsgs, "Failed  wrong number of messages expecting: %d, received: %d", expectingMsgs, receivedMsgs) {
				t.Fatal()
			}
			done = true
		}
	}
	fmt.Println("Done")
}

func TestBadSubscriberMessageConfig(t *testing.T) {
	badMsgConfig := types.MessageBusConfig{
		SubscribeHost: types.HostInfo{
			Host: "\\",
		},
	}

	testClient, err := NewZeroMqClient(badMsgConfig)
	require.NoError(t, err)

	_ = testClient.Connect()

	messages := make(chan types.MessageEnvelope)
	topics := []types.TopicChannel{{Topic: "", Messages: messages}}
	messageErrors := make(chan error)

	err = testClient.Subscribe(topics, messageErrors)

	if !assert.NotNil(t, err, "Expecting error from subscriber to ZMQ") {
		t.Fatal()
	}
}

func TestBadPublisherMessageConfig(t *testing.T) {
	badMsgConfig := types.MessageBusConfig{
		PublishHost: types.HostInfo{
			Host: "//",
		},
	}

	testClient, err := NewZeroMqClient(badMsgConfig)
	require.NoError(t, err)

	_ = testClient.Connect()
	defer func() { _ = testClient.Disconnect() }()

	message := types.MessageEnvelope{
		CorrelationID: "123", Payload: []byte("test bytes"),
	}

	topic := "TestTopic"

	err = testClient.Publish(message, topic)

	if !assert.NotNil(t, err, "Expecting error from publish to ZMQ") {
		t.Fatal()
	}
}

func TestDisconnect(t *testing.T) {

	testMsgConfig := types.MessageBusConfig{
		PublishHost: types.HostInfo{
			Host:     "*",
			Port:     5577,
			Protocol: "tcp",
		},
		SubscribeHost: types.HostInfo{
			Host:     "localhost",
			Port:     5577,
			Protocol: "tcp",
		},
	}

	testClient, err := NewZeroMqClient(testMsgConfig)
	require.NoError(t, err)

	_ = testClient.Connect()

	topic := ""

	runPublishSubscribe(t, testClient, topic, "")

	err = testClient.Disconnect()

	if !assert.NoError(t, err, "Disconnect failed") {
		t.Fatal()
	}

	if !assert.True(t, allSubscribersDisconnected(testClient), "SubscribeStrategy not closed") {
		t.Fatal()
	}

	err = <-testClient.messageErrors
	if !assert.Nil(t, err, "Message error channel is not closed") {
		t.Fatal()
	}

	msgEnvelop := <-testClient.subscribers[0].topic.Messages
	assert.Nil(t, msgEnvelop.Payload, "topic channel is not closed")
}

func allSubscribersDisconnected(client *zeromqClient) bool {
	client.lock.Lock()
	defer client.lock.Unlock()

	allDisconnected := true
	for _, subscriber := range client.subscribers {
		if subscriber.connection.String() != "Socket(CLOSED)" {
			allDisconnected = false
			break
		}
	}
	return allDisconnected
}
