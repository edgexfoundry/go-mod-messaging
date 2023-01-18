/********************************************************************************
 *  Copyright 2020 Dell Inc.
 *  Copyright (c) 2023 Intel Corporation
 *  Copyright (C) 2023 IOTech Ltd
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

// redis package contains tests for the functionality provided in the client.go file.
//
// NOTE: The structure of these tests rely on mocking provided by the Mockery library. Mockery creates mock
// implementations for a specified interface, in this case RedisClient. The Mockery generated code can be found in the
// 'mocks/RedisClient.go'. The generated code was created by using the mockery command line tool provided by Mockery.
// For example execute the following CLI command:
// '$ mockery -name=RedisClient -recursive'
// For more details visit the Mockery Github page https://github.com/vektra/mockery
package redis

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v3/internal/pkg"
	redisMocks "github.com/edgexfoundry/go-mod-messaging/v3/internal/pkg/redis/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var HostInfo = types.HostInfo{
	Host:     "localhost",
	Port:     6379,
	Protocol: "redis",
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name             string
		messageBusConfig types.MessageBusConfig
		wantErr          bool
	}{
		{
			name:             "Successfully create client",
			messageBusConfig: types.MessageBusConfig{},
			wantErr:          false,
		},

		{
			name: "Successfully create client with optional configuration",
			messageBusConfig: types.MessageBusConfig{
				Broker: HostInfo,
				Optional: map[string]string{
					pkg.Password: "Password",
				},
			},
			wantErr: false,
		},

		{
			name: "Invalid Redis Server",
			messageBusConfig: types.MessageBusConfig{
				Broker: types.HostInfo{
					Host:     "!@#$",
					Port:     -1,
					Protocol: "!@#",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.messageBusConfig)
			if tt.wantErr {
				require.Error(t, err)
			}

			if !tt.wantErr {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewClientWithCreator(t *testing.T) {
	tests := []struct {
		name             string
		messageBusConfig types.MessageBusConfig
		creator          RedisClientCreator
		pairCreator      pkg.X509KeyPairCreator
		keyLoader        pkg.X509KeyLoader
		caCertCreator    pkg.X509CaCertCreator
		caCertLoader     pkg.X509CaCertLoader
		pemDecoder       pkg.PEMDecoder
		wantErr          bool
	}{
		{
			name:             "Client with Publish Broker",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo},
			creator:          mockRedisClientCreator(nil, nil),
			pairCreator:      mockCertCreator(nil),
			keyLoader:        mockCertLoader(nil),
			wantErr:          false,
		},
		{
			name:             "Create publisher error",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo},
			creator:          mockRedisClientCreator(nil, errors.New("test error")),
			pairCreator:      mockCertCreator(nil),
			keyLoader:        mockCertLoader(nil),
			wantErr:          true,
		},
		{
			name:             "Client with Subscribe Broker",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo},
			creator:          mockRedisClientCreator(nil, nil),
			pairCreator:      mockCertCreator(nil),
			keyLoader:        mockCertLoader(nil),
			wantErr:          false,
		},
		{
			name:             "Client subscriber error",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo},
			creator:          mockRedisClientCreator(nil, errors.New("test error")),
			pairCreator:      mockCertCreator(nil),
			keyLoader:        mockCertLoader(nil),
			wantErr:          true,
		},
		{
			name:             "Client with optional configuration",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo, Optional: map[string]string{"Password": "TestPassword"}},
			creator:          mockRedisClientCreator(nil, nil),
			pairCreator:      mockCertCreator(nil),
			keyLoader:        mockCertLoader(nil),
			wantErr:          false,
		},
		{
			name: "Client with valid TLS configuration",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo, Optional: map[string]string{
				"CertFile":     "certFile",
				"KeyFile":      "keyFile",
				"CertPEMBlock": "certPRMBlock",
				"KeyPEMBlock":  "keyPEMBlock",
			}},
			creator:     mockRedisClientCreator(nil, nil),
			pairCreator: mockCertCreator(nil),
			keyLoader:   mockCertLoader(nil),
			wantErr:     false,
		},
		{
			name: "Client with valid TLS configuration (cacert file)",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo, Optional: map[string]string{
				"CaFile": "caCertFile",
			}},
			creator:       mockRedisClientCreator(nil, nil),
			caCertCreator: mockCaCertCreator(nil),
			caCertLoader:  mockCaCertLoader(nil),
			pemDecoder:    mockPemDecoder(&pem.Block{}),
			wantErr:       false,
		},
		{
			name: "Client with valid TLS configuration (cacert PEM block)",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo, Optional: map[string]string{
				"CaPEMBlock": "caCertPEMBlock",
			}},
			creator:       mockRedisClientCreator(nil, nil),
			caCertCreator: mockCaCertCreator(nil),
			pemDecoder:    mockPemDecoder(&pem.Block{}),
			wantErr:       false,
		},
		{
			name: "Client with invalid TLS configuration",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo, Optional: map[string]string{
				"SkipCertVerify": "NotABool",
			}},
			creator:     mockRedisClientCreator(nil, nil),
			pairCreator: mockCertCreator(nil),
			keyLoader:   mockCertLoader(nil),
			wantErr:     true,
		},
		{
			name: "Client TLS creation error",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo, Optional: map[string]string{
				"CertFile":     "certFile",
				"KeyFile":      "keyFile",
				"CertPEMBlock": "certPRMBlock",
				"KeyPEMBlock":  "keyPEMBlock",
			}},
			creator:     mockRedisClientCreator(nil, nil),
			pairCreator: mockCertCreator(errors.New("test error")),
			keyLoader:   mockCertLoader(errors.New("test error")),
			wantErr:     true,
		},
		{
			name: "Client TLS creation error - cacert file not found",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo, Optional: map[string]string{
				"CaFile": "caCertFile",
			}},
			creator:      mockRedisClientCreator(nil, nil),
			caCertLoader: mockCaCertLoader(errors.New("test error")),
			wantErr:      true,
		},
		{
			name: "Client TLS creation error - cacert file without PEM block",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo, Optional: map[string]string{
				"CaFile": "caCertFile",
			}},
			creator:      mockRedisClientCreator(nil, nil),
			caCertLoader: mockCaCertLoader(nil),
			pemDecoder:   mockPemDecoder(nil),
			wantErr:      true,
		},
		{
			name: "Client TLS creation error - invalid cacert",
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo, Optional: map[string]string{
				"CaPEMBlock": "caCertPEMBlock",
			}},
			creator:       mockRedisClientCreator(nil, nil),
			caCertCreator: mockCaCertCreator(errors.New("test error")),
			pemDecoder:    mockPemDecoder(&pem.Block{}),
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClientWithCreator(tt.messageBusConfig, tt.creator, tt.pairCreator, tt.keyLoader,
				tt.caCertCreator, tt.caCertLoader, tt.pemDecoder)
			if tt.wantErr {
				require.Error(t, err)
			}

			if !tt.wantErr {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_Connect(t *testing.T) {
	messageClient, err := NewClient(types.MessageBusConfig{})
	require.NoError(t, err)
	err = messageClient.Connect()
	require.NoError(t, err, "Connect is not expected to return an error")
}

func TestClient_Publish(t *testing.T) {
	ValidMessage := types.MessageEnvelope{
		CorrelationID: "abc",
		Payload:       []byte("Test payload"),
		ContentType:   "application/test",
	}

	emptyMessage := types.MessageEnvelope{}

	Topic := "UnitTestTopic"

	tests := []struct {
		name               string
		redisClientCreator RedisClientCreator
		message            types.MessageEnvelope
		topic              string
		wantErr            bool
		errorType          error
	}{
		{
			name: "Send message",
			redisClientCreator: mockRedisClientCreator([]mockOutline{
				{
					methodName: "Send",
					arg: []interface{}{
						Topic,
						ValidMessage,
					},
					ret: []interface{}{nil},
				},
			}, nil),
			message:   ValidMessage,
			topic:     Topic,
			wantErr:   false,
			errorType: nil,
		},
		{
			name: "Send message with missing components",
			redisClientCreator: mockRedisClientCreator([]mockOutline{
				{
					methodName: "Send",
					arg: []interface{}{
						Topic,
						emptyMessage,
					},
					ret: []interface{}{nil},
				},
			}, nil),
			message:   types.MessageEnvelope{},
			topic:     Topic,
			wantErr:   false,
			errorType: nil,
		},
		{
			name:               "Send message with empty topic",
			redisClientCreator: mockRedisClientCreator(nil, nil),
			message:            types.MessageEnvelope{},
			topic:              "",
			wantErr:            true,
			errorType:          pkg.InvalidTopicErr{},
		},
		{
			name:               "Send message with no underlying publishing client",
			redisClientCreator: mockNilRedisClientCreator(),
			message:            types.MessageEnvelope{},
			topic:              Topic,
			wantErr:            true,
			errorType:          pkg.MissingConfigurationErr{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClientWithCreator(types.MessageBusConfig{
				Broker: HostInfo,
			}, tt.redisClientCreator, mockCertCreator(nil), mockCertLoader(nil),
				mockCaCertCreator(nil), mockCaCertLoader(nil), mockPemDecoder(&pem.Block{}))

			require.NoError(t, err)
			err = c.Publish(tt.message, tt.topic)
			if tt.wantErr {
				require.Error(t, err)
				expectedErrorType := reflect.TypeOf(tt.errorType)
				actualErrorType := reflect.TypeOf(err)
				assert.Condition(t, func() (success bool) {
					return actualErrorType.AssignableTo(expectedErrorType)
				}, "Expected error of type %v, but got an error of type %v", expectedErrorType, actualErrorType)
			}

			if !tt.wantErr {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_Subscribe(t *testing.T) {
	tests := []struct {
		name             string
		numberOfMessages int
		numberOfErrors   int
	}{
		{
			name:             "Single message",
			numberOfMessages: 1,
			numberOfErrors:   0,
		},
		{
			name:             "Multiple messages",
			numberOfMessages: 2,
			numberOfErrors:   0,
		},
		{
			name:             "Single error",
			numberOfMessages: 2,
			numberOfErrors:   0,
		},
		{
			name:             "Multiple errors",
			numberOfMessages: 2,
			numberOfErrors:   0,
		},
		{
			name:             "Multiple messages and errors and messages",
			numberOfMessages: 5,
			numberOfErrors:   5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since SubscriptionRedisClientMock doesn't have counters per topic,
			// the numberOfMessages and numberOfErrors will not change depending
			// on the number of topics. That's because the total messages mocked
			// by the Receive method is client specific.
			c, err := NewClientWithCreator(
				types.MessageBusConfig{
					Broker: HostInfo,
				},
				mockSubscriptionClientCreator(tt.numberOfMessages, tt.numberOfErrors),
				mockCertCreator(nil),
				mockCertLoader(nil),
				mockCaCertCreator(nil), mockCaCertLoader(nil),
				mockPemDecoder(&pem.Block{}),
			)
			require.NoError(t, err)

			topics := []types.TopicChannel{
				{
					Topic:    "UnitTestTopic",
					Messages: make(chan types.MessageEnvelope),
				},
				{
					Topic:    "UnitTestTopic2",
					Messages: make(chan types.MessageEnvelope),
				},
			}

			errorMessageChannel := make(chan error)
			err = c.Subscribe(topics, errorMessageChannel)
			require.NoError(t, err)
			readFromChannel(t, topics, tt.numberOfMessages, errorMessageChannel, tt.numberOfErrors)
		})
	}
}

func TestClient_Unsubscribe(t *testing.T) {
	config := types.MessageBusConfig{
		Broker: types.HostInfo{
			Host:     "localhost",
			Port:     6379,
			Protocol: "redis",
		},
	}
	creator := func(redisServerURL string, password string, tlsConfig *tls.Config) (RedisClient, error) {
		redisMock := &redisMocks.RedisClient{}
		redisMock.On("Receive", mock.Anything).After(time.Millisecond*100).Return(types.MessageEnvelope{}, nil)
		return redisMock, nil
	}

	target, err := NewClientWithCreator(config, creator, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	messages := make(chan types.MessageEnvelope, 1)
	errs := make(chan error, 1)

	topics := []types.TopicChannel{
		{
			Topic:    "test1",
			Messages: messages,
		},
		{
			Topic:    "test2",
			Messages: messages,
		},
		{
			Topic:    "test3",
			Messages: messages,
		},
	}

	err = target.Subscribe(topics, errs)
	require.NoError(t, err)

	go func() {
		for {
			select {
			case <-errs:
			case <-messages:
			}
		}
	}()

	exists := target.existingTopics["test1"]
	require.True(t, exists)
	exists = target.existingTopics["test2"]
	require.True(t, exists)
	exists = target.existingTopics["test3"]
	require.True(t, exists)

	err = target.Unsubscribe("test1")
	require.NoError(t, err)

	exists = target.existingTopics["test1"]
	require.False(t, exists)
	exists = target.existingTopics["test2"]
	require.True(t, exists)
	exists = target.existingTopics["test3"]
	require.True(t, exists)

	err = target.Unsubscribe("test1", "test2", "test3")
	require.NoError(t, err)

	exists = target.existingTopics["test1"]
	require.False(t, exists)
	exists = target.existingTopics["test2"]
	require.False(t, exists)
	exists = target.existingTopics["test3"]
	require.False(t, exists)
}

func mockCertCreator(returnError error) pkg.X509KeyPairCreator {
	return func(certPEMBlock []byte, keyPEMBlock []byte) (certificate tls.Certificate, err error) {
		return tls.Certificate{}, returnError
	}
}

func mockCertLoader(returnError error) pkg.X509KeyLoader {
	return func(certFile string, keyFile string) (certificate tls.Certificate, err error) {
		return tls.Certificate{}, returnError
	}
}

func mockCaCertCreator(returnError error) pkg.X509CaCertCreator {
	return func(caCertPEMBlock []byte) (*x509.Certificate, error) {
		return &x509.Certificate{}, returnError
	}
}

func mockCaCertLoader(returnError error) pkg.X509CaCertLoader {
	return func(caCertFile string) ([]byte, error) {
		return []byte{}, returnError
	}
}

func mockPemDecoder(pemBlock *pem.Block) pkg.PEMDecoder {
	return func(data []byte) (*pem.Block, []byte) {
		return pemBlock, nil
	}
}

type mockOutline struct {
	methodName string
	arg        []interface{}
	ret        []interface{}
}

func mockRedisClientCreator(outlines []mockOutline, returnedError error) RedisClientCreator {
	mockRedisClient := &redisMocks.RedisClient{}
	for _, outline := range outlines {
		mockRedisClient.On(outline.methodName, outline.arg...).Return(outline.ret...)
	}

	return func(redisServerURL string, password string, tlsConfig *tls.Config) (RedisClient, error) {
		return mockRedisClient, returnedError
	}
}

func mockNilRedisClientCreator() RedisClientCreator {
	return func(redisServerURL string, password string, tlsConfig *tls.Config) (RedisClient, error) {
		return nil, nil
	}
}

func mockSubscriptionClientCreator(numberOfMessages int, numberOfErrors int) RedisClientCreator {
	return func(redisServerURL string, password string, tlsConfig *tls.Config) (RedisClient, error) {
		return &SubscriptionRedisClientMock{
			NumberOfMessages: numberOfMessages,
			NumberOfErrors:   numberOfErrors,
			counterMutex:     &sync.Mutex{},
		}, nil
	}
}

type SubscriptionRedisClientMock struct {
	NumberOfMessages int
	NumberOfErrors   int

	// Keep track of the entities returned
	messagesReturned int
	errorsReturned   int

	counterMutex *sync.Mutex
}

func (r *SubscriptionRedisClientMock) Send(string, types.MessageEnvelope) error {
	panic("implement me")

}

func (r *SubscriptionRedisClientMock) Receive(topic string) (*types.MessageEnvelope, error) {
	r.counterMutex.Lock()

	if r.messagesReturned < r.NumberOfMessages {
		r.messagesReturned++

		defer r.counterMutex.Unlock()
		return createMessage(topic, r.messagesReturned), nil
	}

	if r.errorsReturned < r.NumberOfErrors {
		r.errorsReturned++

		defer r.counterMutex.Unlock()
		return nil, fmt.Errorf("test error %d", r.errorsReturned) // Adding count to make error unique, so it isn't ignored as duplicate.
	}

	r.counterMutex.Unlock()

	for {
		// Sleep to simulate no more messages which will block the caller.
		// This does not affect the execution of the test as this is run in a different Go-routine
		time.Sleep(time.Duration(10) * time.Second)
	}
}

func (r *SubscriptionRedisClientMock) Close() error {
	panic("implement me")
}

func createMessage(topic string, messageNumber int) *types.MessageEnvelope {
	return &types.MessageEnvelope{
		ReceivedTopic: topic,
		CorrelationID: "test",
		Payload:       []byte(fmt.Sprintf("Message #%d", messageNumber)),
		ContentType:   "application/test",
	}
}

func readFromChannel(
	t *testing.T,
	topics []types.TopicChannel,
	expectedNumberOfMessages int,
	errorsChannel <-chan error,
	expectedNumberOfErrors int,
) {

	if len(topics) != 2 {
		panic("Length of input topics should be 2")
	}

	for expectedNumberOfMessages > 0 || expectedNumberOfErrors > 0 {
		select {
		case message := <-topics[0].Messages:
			assert.Equal(t, topics[0].Topic, message.ReceivedTopic, "ReceivedTopic not as expected")
			expectedNumberOfMessages -= 1
			continue

		case message := <-topics[1].Messages:
			assert.Equal(t, topics[1].Topic, message.ReceivedTopic, "ReceivedTopic not as expected")
			expectedNumberOfMessages -= 1
			continue

		case <-errorsChannel:
			expectedNumberOfErrors -= 1
			continue

		case <-time.After(10 * time.Second):
			t.Fatal("Timed out waiting for messages")
		}
	}

	require.Zero(t, expectedNumberOfMessages, "Too many messages")
	require.Zero(t, expectedNumberOfErrors, "Too many errors")
}

func TestClient_Disconnect(t *testing.T) {
	tests := []struct {
		name               string
		redisClientCreator RedisClientCreator
		messageBusConfig   types.MessageBusConfig
		wantErr            bool
	}{
		{
			name: "Close publisher",
			redisClientCreator: mockRedisClientCreator([]mockOutline{
				{
					methodName: "Close",
					arg:        nil,
					ret:        []interface{}{nil},
				}}, nil),
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo},
			wantErr:          false,
		},
		{
			name: "Close publisher error",
			redisClientCreator: mockRedisClientCreator([]mockOutline{
				{
					methodName: "Close",
					arg:        nil,
					ret:        []interface{}{errors.New("test error")},
				}}, nil),
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo},
			wantErr:          true,
		},
		{
			name: "Close subscriber",
			redisClientCreator: mockRedisClientCreator([]mockOutline{
				{
					methodName: "Close",
					arg:        nil,
					ret:        []interface{}{nil},
				}}, nil),
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo},
			wantErr:          false,
		},
		{
			name: "Close subscriber error",
			redisClientCreator: mockRedisClientCreator([]mockOutline{
				{
					methodName: "Close",
					arg:        nil,
					ret:        []interface{}{errors.New("test error")},
				}}, nil),
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo},
			wantErr:          true,
		},
		{
			name: "Close publisher and subscriber",
			redisClientCreator: mockRedisClientCreator([]mockOutline{
				{
					methodName: "Close",
					arg:        nil,
					ret:        []interface{}{nil},
				}}, nil),
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo},
			wantErr:          false,
		},
		{
			name: "Close publisher and subscriber error",
			redisClientCreator: mockRedisClientCreator([]mockOutline{
				{
					methodName: "Close",
					arg:        nil,
					ret:        []interface{}{errors.New("test error")},
				}}, nil),
			messageBusConfig: types.MessageBusConfig{Broker: HostInfo},
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClientWithCreator(tt.messageBusConfig, tt.redisClientCreator, mockCertCreator(nil),
				mockCertLoader(nil), mockCaCertCreator(nil), mockCaCertLoader(nil),
				mockPemDecoder(&pem.Block{}))
			require.NoError(t, err)
			err = c.Disconnect()
			if tt.wantErr {
				require.Error(t, err)
			}

			if !tt.wantErr {
				require.NoError(t, err)
			}
		})
	}
}

func TestConvertToRedisTopicScheme(t *testing.T) {
	tests := []struct {
		name          string
		topic         string
		expectedTopic string
	}{
		{"topic with separator", "test/UnitTestTopic", "test.UnitTestTopic"},
		{"topic with multi level wildcard", "test/UnitTestTopic/#", "test.UnitTestTopic.*"},
		{"topic with single level wildcard", "test/+/UnitTestTopic", "test.*.UnitTestTopic"},
		{"topic with mixed wildcards", "test/+/UnitTestTopic/#", "test.*.UnitTestTopic.*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisTopic := convertToRedisTopicScheme(tt.topic)
			assert.Equal(t, redisTopic, tt.expectedTopic)
		})
	}
}

func TestConvertFromRedisTopicScheme(t *testing.T) {
	tests := []struct {
		name          string
		topic         string
		expectedTopic string
	}{
		{"topic with separator", "test.UnitTestTopic", "test/UnitTestTopic"},
		{"topic with wildcard at the end", "test.UnitTestTopic.*", "test/UnitTestTopic/#"},
		{"topic with wildcard in the middle", "test.*.UnitTestTopic", "test/+/UnitTestTopic"},
		{"topic with multiple wildcards", "test.*.UnitTestTopic.*", "test/+/UnitTestTopic/#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mqttTopic := convertFromRedisTopicScheme(tt.topic)
			assert.Equal(t, mqttTopic, tt.expectedTopic)
		})
	}
}
