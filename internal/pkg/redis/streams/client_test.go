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

// streams package contains tests for the functionality provided in the client.go file.
//
// NOTE: The structure of these tests rely on mocking provided by the Mockery library. Mockery creates mock
// implementations for a specified interface, in this case RedisClient. The Mockery generated code can be found in the
// 'mocks/RedisClient.go'. The generated code was created by using the mockery command line tool provided by Mockery.
// For example execute the following CLI command:
// '$ mockery -name=RedisClient -recursive'
// For more details visit the Mockery Github page https://github.com/vektra/mockery
package streams

import (
	"crypto/tls"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v2/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/v2/internal/pkg/redis/streams/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging/redis"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

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
				PublishHost: HostInfo,
				Optional: map[string]string{
					redis.Password: "Password",
				},
			},
			wantErr: false,
		},

		{
			name: "Invalid Redis Server",
			messageBusConfig: types.MessageBusConfig{
				PublishHost: types.HostInfo{
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
		wantErr          bool
	}{
		{
			name:             "Client with Publish Host",
			messageBusConfig: types.MessageBusConfig{PublishHost: HostInfo},
			creator:          mockRedisClientCreator(nil, nil),
			pairCreator:      mockCertCreator(nil),
			keyLoader:        mockCertLoader(nil),
			wantErr:          false,
		},
		{
			name:             "Create publisher error",
			messageBusConfig: types.MessageBusConfig{PublishHost: HostInfo},
			creator:          mockRedisClientCreator(nil, errors.New("test error")),
			pairCreator:      mockCertCreator(nil),
			keyLoader:        mockCertLoader(nil),
			wantErr:          true,
		},
		{
			name:             "Client with Subscribe Host",
			messageBusConfig: types.MessageBusConfig{SubscribeHost: HostInfo},
			creator:          mockRedisClientCreator(nil, nil),
			pairCreator:      mockCertCreator(nil),
			keyLoader:        mockCertLoader(nil),
			wantErr:          false,
		},
		{
			name:             "Client subscriber error",
			messageBusConfig: types.MessageBusConfig{SubscribeHost: HostInfo},
			creator:          mockRedisClientCreator(nil, errors.New("test error")),
			pairCreator:      mockCertCreator(nil),
			keyLoader:        mockCertLoader(nil),
			wantErr:          true,
		},

		{
			name:             "Client with optional configuration",
			messageBusConfig: types.MessageBusConfig{SubscribeHost: HostInfo, Optional: map[string]string{"Password": "TestPassword"}},
			creator:          mockRedisClientCreator(nil, nil),
			pairCreator:      mockCertCreator(nil),
			keyLoader:        mockCertLoader(nil),
			wantErr:          false,
		},
		{
			name: "Client with valid TLS configuration",
			messageBusConfig: types.MessageBusConfig{SubscribeHost: HostInfo, Optional: map[string]string{
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
			name: "Client with invalid TLS configuration",
			messageBusConfig: types.MessageBusConfig{SubscribeHost: HostInfo, Optional: map[string]string{
				"SkipCertVerify": "NotABool",
			}},
			creator:     mockRedisClientCreator(nil, nil),
			pairCreator: mockCertCreator(nil),
			keyLoader:   mockCertLoader(nil),
			wantErr:     true,
		},
		{
			name: "Client TLS creation error",
			messageBusConfig: types.MessageBusConfig{SubscribeHost: HostInfo, Optional: map[string]string{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClientWithCreator(tt.messageBusConfig, tt.creator, tt.pairCreator, tt.keyLoader)
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
		Checksum:      "123",
		CorrelationID: "abc",
		Payload:       []byte("Test payload"),
		ContentType:   "application/test",
	}

	MessageValues := map[string]interface{}{
		"CorrelationID": ValidMessage.CorrelationID,
		"Payload":       ValidMessage.Payload,
		"Checksum":      ValidMessage.Checksum,
		"ContentType":   ValidMessage.ContentType,
	}

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
					methodName: "AddToStream",
					arg: []interface{}{
						fmt.Sprintf(GoModMessagingNamespaceFormat, "UnitTestTopic"),
						MessageValues,
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
					methodName: "AddToStream",
					arg: []interface{}{
						fmt.Sprintf(GoModMessagingNamespaceFormat, "UnitTestTopic"),
						map[string]interface{}{
							"CorrelationID": "",
							"Payload":       []uint8(nil),
							"Checksum":      "",
							"ContentType":   "",
						},
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
				PublishHost: HostInfo,
			}, tt.redisClientCreator, mockCertCreator(nil), mockCertLoader(nil))

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
	Topic := "UnitTestTopic"
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
			c, err := NewClientWithCreator(types.MessageBusConfig{
				SubscribeHost: HostInfo,
			},
				mockSubscriptionClientCreator(tt.numberOfMessages, tt.numberOfErrors),
				mockCertCreator(nil),
				mockCertLoader(nil))

			require.NoError(t, err)
			messageChannel := make(chan types.MessageEnvelope)
			topics := []types.TopicChannel{
				{
					Topic:    Topic,
					Messages: messageChannel,
				},
			}

			errorMessageChannel := make(chan error)
			err = c.Subscribe(topics, errorMessageChannel)
			require.NoError(t, err)
			receivedExpectedMessagesWaitGroup := &sync.WaitGroup{}
			receivedExpectedMessagesWaitGroup.Add(1)
			readFromChannel(messageChannel, tt.numberOfMessages, errorMessageChannel, tt.numberOfErrors, receivedExpectedMessagesWaitGroup)
			receivedExpectedMessagesWaitGroup.Wait()
		})
	}
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

type mockOutline struct {
	methodName string
	arg        []interface{}
	ret        []interface{}
}

func mockRedisClientCreator(outlines []mockOutline, returnedError error) RedisClientCreator {
	mockRedisClient := &mocks.RedisClient{}
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
		}, nil
	}
}

type SubscriptionRedisClientMock struct {
	NumberOfMessages int
	NumberOfErrors   int

	// Keep track of the entities returned
	messagesReturned int
	errorsReturned   int
}

func (r *SubscriptionRedisClientMock) AddToStream(string, map[string]interface{}) error {
	panic("implement me")

}

func (r *SubscriptionRedisClientMock) ReadFromStream(string) ([]types.MessageEnvelope, error) {
	if r.messagesReturned < r.NumberOfMessages {
		r.messagesReturned = r.NumberOfMessages
		return createMessages(r.NumberOfMessages), nil
	}

	if r.errorsReturned < r.NumberOfErrors {
		r.errorsReturned += 1
		return nil, errors.New("test error")
	}

	for {
		// Sleep to simulate no more messages which will block the caller.
		// This does not affect the execution of the test as this is run in a different Go-routine
		time.Sleep(time.Duration(10) * time.Second)
	}
}

func (r *SubscriptionRedisClientMock) Close() error {
	panic("implement me")
}
func createMessages(numberOfMessages int) []types.MessageEnvelope {
	messages := make([]types.MessageEnvelope, numberOfMessages)
	for index := 0; index < numberOfMessages; index++ {
		messages[index] = types.MessageEnvelope{
			Checksum:      strconv.FormatInt(int64(index), 10),
			CorrelationID: "test",
			Payload:       []byte(fmt.Sprintf("Message #%d", index)),
			ContentType:   "application/test",
		}
	}

	return messages
}

func readFromChannel(
	messageChannel <-chan types.MessageEnvelope,
	expectedNumberOfMessages int,
	errorsChannel <-chan error,
	expectedNumberOfErrors int,
	group *sync.WaitGroup) {

	for expectedNumberOfMessages > 0 || expectedNumberOfErrors > 0 {
		select {
		case <-messageChannel:
			{
				expectedNumberOfMessages -= 1
				continue
			}

		case <-errorsChannel:
			{
				expectedNumberOfErrors -= 1
				continue
			}
		}
	}

	group.Done()
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
			messageBusConfig: types.MessageBusConfig{PublishHost: HostInfo},
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
			messageBusConfig: types.MessageBusConfig{PublishHost: HostInfo},
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
			messageBusConfig: types.MessageBusConfig{SubscribeHost: HostInfo},
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
			messageBusConfig: types.MessageBusConfig{SubscribeHost: HostInfo},
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
			messageBusConfig: types.MessageBusConfig{PublishHost: HostInfo, SubscribeHost: HostInfo},
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
			messageBusConfig: types.MessageBusConfig{PublishHost: HostInfo, SubscribeHost: HostInfo},
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClientWithCreator(tt.messageBusConfig, tt.redisClientCreator, mockCertCreator(nil), mockCertLoader(nil))
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
