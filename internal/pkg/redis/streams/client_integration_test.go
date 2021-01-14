// +build redisIntegration

/********************************************************************************
 *  Copyright 2020 Dell Inc.
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
 * - The Redis server URL set as the environment variable REDIS_SERVER_TEST. Otherwise the default
 * 'redis://localhost:6379' will be used
 */
package streams

import (
	"net/url"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

const (
	RedisURLEnvName = "REDIS_SERVER_TEST"
	DefaultRedisURL = "redis://localhost:6379"
	TestStream      = "IntegrationTest"
)

func TestRedisStreamsClientIntegration(t *testing.T) {
	redisHostInfo := getRedisHostInfo(t)
	client, err := NewClient(types.MessageBusConfig{
		PublishHost:   redisHostInfo,
		SubscribeHost: redisHostInfo,
	})

	require.NoError(t, err, "Failed to create Redis client")
	testMessage := types.MessageEnvelope{
		Checksum:      "abcde",
		CorrelationID: "12345",
		Payload:       []byte("test-message")}

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
