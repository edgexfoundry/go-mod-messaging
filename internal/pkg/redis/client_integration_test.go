//go:build redisIntegration

/********************************************************************************
 *  Copyright 2020 Dell Inc.
 *  Copyright (c) 2023 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

/**
 * This test file contains integration tests for the RedisStreams Client. These integration tests are disabled by default
 * as they require additional external resources to be executed properly. These tests are enabled by providing the
 * `redisIntegration` tag when running the tests. For example 'go test ./... -tags=redisIntegration`.
 *
 * This is a list of the requirements necessary to run the tests in this file:
 * - Redis server with no password required by clients.
 * - The Redis server URL set as the environment variable REDIS_SERVER_TEST, Otherwise the default
 * 'redis://localhost:6379' will be used
 */
package redis

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
)

const (
	RedisURLEnvName = "REDIS_SERVER_TEST"
	DefaultRedisURL = "redis://localhost:6379"
	TestStream      = "IntegrationTest"
)

func TestRedisClientIntegration(t *testing.T) {
	redisHostInfo := getRedisHostInfo(t)
	client, err := NewClient(types.MessageBusConfig{
		Broker: redisHostInfo,
	})

	require.NoError(t, err, "Failed to create Redis client")
	testMessage := types.MessageEnvelope{
		CorrelationID: "12345",
		Payload:       []byte("test-message"),
		ReceivedTopic: "IntegrationTest",
	}

	testCompleteWaitGroup := &sync.WaitGroup{}
	testCompleteWaitGroup.Add(1)
	createSub(t, &client, TestStream, testMessage, testCompleteWaitGroup)
	// TODO Add proper hook/wait
	time.Sleep(time.Duration(1) * time.Second)
	err = client.Publish(testMessage, TestStream)
	require.NoError(t, err, "Failed to publish test message")
	testCompleteWaitGroup.Wait()

	err = client.Disconnect()
	require.NoError(t, err)
}

// TestRedisUnsubscribeIntegration end-to-end test of subscribing and unsubscribing.
// Redis must be running with Device virtual publishing events.
func TestRedisUnsubscribeIntegration(t *testing.T) {
	redisHostInfo := getRedisHostInfo(t)
	client, err := NewClient(types.MessageBusConfig{
		Broker: redisHostInfo,
	})

	require.NoError(t, err, "Failed to create Redis client")

	messages := make(chan types.MessageEnvelope, 1)
	errs := make(chan error, 1)

	eventTopic := "edgex/events/#"
	topics := []types.TopicChannel{
		{
			Topic:    eventTopic,
			Messages: messages,
		},
	}

	println("Subscribing to topic: " + eventTopic)
	err = client.Subscribe(topics, errs)
	require.NoError(t, err)
	require.Equal(t, 1, len(client.existingTopics))

	messageCount := 0

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(time.Second * 10):
				return

			case err = <-errs:
				require.Failf(t, "failed", "Unexpected error message received: %v", err)
				return

			case message := <-messages:
				println(fmt.Sprintf("Received message from topic: %v", message.ReceivedTopic))
				messageCount++
				if messageCount > 3 {
					println("Unsubscribing from topic: " + eventTopic)
					err = client.Unsubscribe(eventTopic)
					require.NoError(t, err)
				}
			}
		}

	}()

	wg.Wait()
	assert.Greater(t, messageCount, 3)
	assert.Equal(t, 0, len(client.existingTopics))
}

func TestRedisRequestIntegration(t *testing.T) {
	redisHostInfo := getRedisHostInfo(t)
	client, err := NewClient(types.MessageBusConfig{
		Broker:              redisHostInfo,
		ResponseTopicPrefix: "edgex/response",
	})

	require.NoError(t, err, "Failed to create Redis client")

	messages := make(chan types.MessageEnvelope, 1)
	errs := make(chan error, 1)

	serviceName := "test-service"
	requestTopic := "edgex/request"
	topics := []types.TopicChannel{
		{
			Topic:    requestTopic,
			Messages: messages,
		},
	}

	println("Subscribing to topic: " + requestTopic)
	err = client.Subscribe(topics, errs)
	require.NoError(t, err)
	require.Equal(t, 1, len(client.existingTopics))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(time.Second * 10):
				return

			case err = <-errs:
				require.Failf(t, "failed", "Unexpected error message received: %v", err)
				return

			case message := <-messages:
				println(fmt.Sprintf("Received message from topic: %v", message.ReceivedTopic))

				responseTopic := strings.Join([]string{client.responseTopicPrefix, serviceName, message.RequestID}, "/")
				println(fmt.Sprintf("Publishing response message on topic: %v", responseTopic))

				err = client.Publish(types.MessageEnvelope{RequestID: message.RequestID}, responseTopic)
				require.NoError(t, err)
				return
			}
		}
	}()

	time.Sleep(time.Second)
	requestId := uuid.NewString()
	println(fmt.Sprintf("Sending request to topic %s with requestId %s", requestTopic, requestId))
	response, err := client.Request(types.MessageEnvelope{RequestID: requestId}, serviceName, requestTopic, time.Second*10)
	require.NoError(t, err)
	assert.Equal(t, requestId, response.RequestID)
}

func createSub(t *testing.T, client *Client, stream string, expectedMessage types.MessageEnvelope, doneWaitGroup *sync.WaitGroup) {
	msgError := make(chan error)
	messageChannel := make(chan types.MessageEnvelope)
	err := client.Subscribe([]types.TopicChannel{
		{
			Topic:    stream,
			Messages: messageChannel,
		},
	}, msgError)

	require.NoError(t, err, "Failed to create subscription")
	go func() {
		defer doneWaitGroup.Done()
		select {
		case message := <-msgError:
			assert.Nil(t, message, "Unexpected error message received: %v", message)

		case message := <-messageChannel:
			assert.Equal(t, expectedMessage, message)
		}
	}()
}

func getRedisHostInfo(t *testing.T) types.HostInfo {
	redisURLString := getRedisURL()
	redisURL, err := url.Parse(redisURLString)
	if err != nil {
		require.NoError(t, err, "Unable to parse the Redis URL %s: %v", redisURL, err)
		t.FailNow()
	}

	port, err := strconv.Atoi(redisURL.Port())
	if err != nil {
		require.NoError(t, err)
		t.FailNow()
	}

	host := redisURL.Hostname()
	scheme := redisURL.Scheme

	return types.HostInfo{
		Host:     host,
		Port:     port,
		Protocol: scheme,
	}
}

func getRedisURL() string {
	redisURLString := os.Getenv(RedisURLEnvName)
	if redisURLString == "" {
		return DefaultRedisURL
	}

	return redisURLString
}
