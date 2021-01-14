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

package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v2/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging/mqtt"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	pahoMqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

var OptionalPropertiesNoTls = map[string]string{
	mqtt.Username:  "TestUser",
	mqtt.Password:  "TestPassword",
	mqtt.ClientId:  "TestClientID",
	mqtt.Qos:       "1",
	mqtt.KeepAlive: "3",
	mqtt.Retained:  "true",
}

var OptionalPropertiesCertCreate = map[string]string{
	mqtt.Username:       "TestUser",
	mqtt.Password:       "TestPassword",
	mqtt.ClientId:       "TestClientID",
	mqtt.Qos:            "1",
	mqtt.KeepAlive:      "3",
	mqtt.Retained:       "true",
	mqtt.CertPEMBlock:   "CertBytes",
	mqtt.KeyPEMBlock:    "KeyBytes",
	mqtt.ConnectTimeout: "1",
}

var OptionalPropertiesCertLoad = map[string]string{
	mqtt.Username:       "TestUser",
	mqtt.Password:       "TestPassword",
	mqtt.ClientId:       "TestClientID",
	mqtt.Qos:            "1",
	mqtt.KeepAlive:      "3",
	mqtt.Retained:       "true",
	mqtt.CertFile:       "./cert",
	mqtt.KeyFile:        "./key",
	mqtt.ConnectTimeout: "1",
}

var TcpHostInfo = types.HostInfo{Host: "localhost", Protocol: "tcp", Port: 1883}
var TlsHostInfo = types.HostInfo{Host: "localhost", Protocol: "tls", Port: 8883}
var TcpsHostInfo = types.HostInfo{Host: "localhost", Protocol: "tcps", Port: 8883}
var SslHostInfo = types.HostInfo{Host: "localhost", Protocol: "ssl", Port: 8883}

// TestMessageBusConfig defines a simple configuration used for testing successful options parsing.
var TestMessageBusConfig = types.MessageBusConfig{
	PublishHost: TcpsHostInfo,
	Optional:    OptionalPropertiesNoTls,
}
var TestMessageBusConfigTlsCreate = types.MessageBusConfig{
	PublishHost: TlsHostInfo,
	Optional:    OptionalPropertiesCertCreate,
}

var TestMessageBusConfigTlsLoad = types.MessageBusConfig{
	PublishHost: TlsHostInfo,
	Optional:    OptionalPropertiesCertLoad,
}

// MockToken implements Token and gives control over the information returned to the caller of the various
// Client methods, such as Connect.
type MockToken struct {
	waitTimeOut bool
	err         error
}

// SuccessfulMockToken creates a MockToken which returns data indicating a successfully completed operation.
func SuccessfulMockToken() MockToken {
	return MockToken{
		waitTimeOut: true,
		err:         nil,
	}
}

// TimeoutNoErrorMockToken creates a MockToken which returns data indicating a timeout has occurred and does not provide
// an error with additional information.
func TimeoutNoErrorMockToken() MockToken {
	return MockToken{
		waitTimeOut: false,
		err:         nil,
	}
}

// TimeoutWithErrorMockToken creates a MockToken which returns data indicating a timeout has occurred and provides an
// error with additional information.
func TimeoutWithErrorMockToken() MockToken {
	return MockToken{
		waitTimeOut: false,
		err:         errors.New("timeout while trying to complete operation"),
	}
}

// ErrorMockToken creates a MockToken which returns data indicating an error has occurred by providing an error.
func ErrorMockToken() MockToken {
	return MockToken{
		waitTimeOut: true,
		err:         errors.New("operation failed"),
	}
}

func (mt MockToken) WaitTimeout(time.Duration) bool {
	return mt.waitTimeOut
}

func (mt MockToken) Error() error {
	return mt.err
}

// MockMQTTClient implements the Client interface and allows for control over the returned data when invoking it's
// methods.
type MockMQTTClient struct {
	subscriptions map[string]pahoMqtt.MessageHandler
	// MockTokens used to control the returned values for the respective functions.
	connect   MockToken
	publish   MockToken
	subscribe MockToken
}

func (mc MockMQTTClient) Connect() pahoMqtt.Token {
	return &mc.connect
}

func (mc MockMQTTClient) Publish(topic string, _ byte, _ bool, message interface{}) pahoMqtt.Token {
	handler, ok := mc.subscriptions[topic]
	if !ok {
		return &mc.publish
	}

	go handler(mc, MockMessage{payload: message.([]byte)})
	return &mc.publish
}

func (mc MockMQTTClient) Subscribe(topic string, _ byte, handler pahoMqtt.MessageHandler) pahoMqtt.Token {
	mc.subscriptions[topic] = handler
	return &mc.subscribe
}

func (mc MockMQTTClient) Disconnect(uint) {
	// No implementation required.
}

func (mt MockToken) Wait() bool {
	panic("function not expected to be invoked")
}

func (MockMQTTClient) IsConnected() bool {
	return false
}

func (MockMQTTClient) IsConnectionOpen() bool {
	panic("function not expected to be invoked")
}

func (MockMQTTClient) SubscribeMultiple(map[string]byte, pahoMqtt.MessageHandler) pahoMqtt.Token {
	panic("function not expected to be invoked")
}

func (MockMQTTClient) Unsubscribe(...string) pahoMqtt.Token {
	panic("function not expected to be invoked")
}

func (MockMQTTClient) AddRoute(string, pahoMqtt.MessageHandler) {
	panic("function not expected to be invoked")
}

func (MockMQTTClient) OptionsReader() pahoMqtt.ClientOptionsReader {
	return pahoMqtt.NewClient(pahoMqtt.NewClientOptions()).OptionsReader()

}

// MockMessage implements the Message interface and allows for control over the returned data when a MessageHandler is
// invoked.
type MockMessage struct {
	payload []byte
}

func (mm MockMessage) Payload() []byte {
	return mm.payload
}

func (MockMessage) Duplicate() bool {
	panic("function not expected to be invoked")
}

func (MockMessage) Qos() byte {
	panic("function not expected to be invoked")
}

func (MockMessage) Retained() bool {
	panic("function not expected to be invoked")
}

func (MockMessage) Topic() string {
	panic("function not expected to be invoked")
}

func (MockMessage) MessageID() uint16 {
	panic("function not expected to be invoked")
}

func (MockMessage) Ack() {
	panic("function not expected to be invoked")
}

// mockClientCreator higher-order function which creates a function that constructs a MockMQTTClient
func mockClientCreator(connect MockToken, publish MockToken, subscribe MockToken) ClientCreator {
	return func(config types.MessageBusConfig, handler pahoMqtt.OnConnectHandler) (pahoMqtt.Client, error) {
		return MockMQTTClient{
			connect:       connect,
			publish:       publish,
			subscribe:     subscribe,
			subscriptions: make(map[string]pahoMqtt.MessageHandler),
		}, nil

	}
}

func TestInvalidClientOptions(t *testing.T) {
	invalidOptions := types.MessageBusConfig{PublishHost: types.HostInfo{
		Host:     "    ",
		Port:     0,
		Protocol: "    ",
	}}

	client, _ := NewMQTTClient(invalidOptions)
	err := client.Connect()
	require.Error(t, err)
}

func TestInvalidTlsOptions(t *testing.T) {
	options := types.MessageBusConfig{
		PublishHost: TlsHostInfo,
		Optional: map[string]string{
			"CertFile": "./does-not-exist",
			"KeyFile":  "./does-not-exist",
		},
	}
	client, _ := NewMQTTClient(options)
	err := client.Connect()
	require.Error(t, err)
}

func TestClientCreatorTLS(t *testing.T) {
	tests := []struct {
		name            string
		hostConfig      types.HostInfo
		optionalConfig  map[string]string
		certCreator     pkg.X509KeyPairCreator
		certLoader      pkg.X509KeyLoader
		expectError     bool
		expectTLSConfig bool
	}{
		{
			name:       "Create TLS Config from PEM Block",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				mqtt.CertPEMBlock:   "CertPEMBlock",
				mqtt.KeyPEMBlock:    "KeyPEMBlock",
				mqtt.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Create TCPS Config from PEM Block",
			hostConfig: TcpsHostInfo,
			optionalConfig: map[string]string{
				mqtt.CertPEMBlock:   "CertPEMBlock",
				mqtt.KeyPEMBlock:    "KeyPEMBlock",
				mqtt.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Create SSL Config from PEM Block",
			hostConfig: SslHostInfo,
			optionalConfig: map[string]string{
				mqtt.CertPEMBlock:   "CertPEMBlock",
				mqtt.KeyPEMBlock:    "KeyPEMBlock",
				mqtt.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Skip TLS Config from PEM Block for non-supported TLS protocols",
			hostConfig: TcpHostInfo,
			optionalConfig: map[string]string{
				mqtt.CertPEMBlock:   "CertPEMBlock",
				mqtt.KeyPEMBlock:    "KeyPEMBlock",
				mqtt.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: false,
		},
		{
			name:       "Fail Create TLS Config from PEM File",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				mqtt.CertPEMBlock:   "CertPEMBlock",
				mqtt.KeyPEMBlock:    "KeyPEMBlock",
				mqtt.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(errors.New("test error")),
			certLoader:      mockCertLoader(nil),
			expectError:     true,
			expectTLSConfig: false,
		},
		{
			name:       "Load TLS Config from Cert File",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				mqtt.CertFile:       "./cert",
				mqtt.KeyFile:        "./key",
				mqtt.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Load TCPS Config from Cert File",
			hostConfig: TcpsHostInfo,
			optionalConfig: map[string]string{
				mqtt.CertFile:       "./cert",
				mqtt.KeyFile:        "./key",
				mqtt.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Load SSL Config from Cert File",
			hostConfig: SslHostInfo,
			optionalConfig: map[string]string{
				mqtt.CertFile:       "./cert",
				mqtt.KeyFile:        "./key",
				mqtt.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Skip Load TLS Config from Cert File for un-supported protocols",
			hostConfig: TcpHostInfo,
			optionalConfig: map[string]string{
				mqtt.CertFile:       "./cert",
				mqtt.KeyFile:        "./key",
				mqtt.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: false,
		},
		{
			name:       "Fail Load TLS Config from Cert File",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				mqtt.CertFile:       "./cert",
				mqtt.KeyFile:        "./key",
				mqtt.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(errors.New("test error")),
			expectError:     true,
			expectTLSConfig: false,
		},
		{
			name:       "Fail Load TLS Config For Invalid Options",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				mqtt.Qos:            "abc",
				mqtt.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(errors.New("test error")),
			expectError:     true,
			expectTLSConfig: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _ := NewMQTTClientWithCreator(
				types.MessageBusConfig{
					PublishHost: test.hostConfig,
					Optional:    test.optionalConfig,
				},
				json.Marshal,
				json.Unmarshal,
				ClientCreatorWithCertLoader(test.certCreator, test.certLoader))

			err := client.Connect()

			// Expecting a timeout error since creating mqtt client now at the beginning of the Connect() function
			_, ok := err.(TimeoutErr)
			if ok {
				err = nil
			}

			if test.expectError {
				require.Error(t, err)
				return // End test for expected error
			} else {
				require.NoError(t, err)
			}

			clientOptions := client.mqttClient.OptionsReader()
			tlsConfig := clientOptions.TLSConfig()
			if test.expectTLSConfig {
				assert.NotNil(t, tlsConfig, "Failed to configure TLS for underlying client")
			} else {
				assert.Nil(t, tlsConfig, "Expected TLS configuration to be not be provided.")
			}
		})
	}
}

func TestClientCreatorTlsLoader(t *testing.T) {
	client, _ := NewMQTTClientWithCreator(
		TestMessageBusConfigTlsLoad,
		json.Marshal,
		json.Unmarshal,
		ClientCreatorWithCertLoader(mockCertCreator(nil), mockCertLoader(nil)))

	err := client.Connect()
	// Expecting a timeout error since creating mqtt client now at the beginning of the Connect() function
	_, ok := err.(TimeoutErr)
	if ok {
		err = nil
	}

	require.NoError(t, err)
	clientOptions := client.mqttClient.OptionsReader()
	tlsConfig := clientOptions.TLSConfig()
	assert.NotNil(t, tlsConfig, "Failed to configure TLS for underlying client")
}

func TestClientCreatorTlsLoadError(t *testing.T) {
	client, _ := NewMQTTClientWithCreator(
		TestMessageBusConfigTlsLoad,
		json.Marshal,
		json.Unmarshal,
		ClientCreatorWithCertLoader(mockCertCreator(nil), mockCertLoader(errors.New("test error"))))

	err := client.Connect()
	// Expecting a timeout error since creating mqtt client now at the beginning of the Connect() function
	_, ok := err.(TimeoutErr)
	if ok {
		err = nil
	}

	assert.Error(t, err, "Expected error for invalid CertFile and KeyFile file locations")
}

func TestClientCreatorTlsCreator(t *testing.T) {
	client, _ := NewMQTTClientWithCreator(
		TestMessageBusConfigTlsCreate,
		json.Marshal,
		json.Unmarshal,
		ClientCreatorWithCertLoader(mockCertCreator(nil), mockCertLoader(nil)))

	err := client.Connect()
	// Expecting a timeout error since creating mqtt client now at the beginning of the Connect() function
	_, ok := err.(TimeoutErr)
	if ok {
		err = nil
	}

	require.NoError(t, err)
	clientOptions := client.mqttClient.OptionsReader()
	tlsConfig := clientOptions.TLSConfig()
	assert.NotNil(t, tlsConfig, "Failed to configure TLS for underlying client")
}

func TestClientCreatorTlsCreatorError(t *testing.T) {
	client, _ := NewMQTTClientWithCreator(
		TestMessageBusConfigTlsCreate,
		json.Marshal,
		json.Unmarshal,
		ClientCreatorWithCertLoader(mockCertCreator(errors.New("test error")), mockCertLoader(nil)))

	err := client.Connect()
	// Expecting a timeout error since creating mqtt client now at the beginning of the Connect() function
	_, ok := err.(TimeoutErr)
	if ok {
		err = nil
	}

	assert.Error(t, err, "Expected error for invalid CertFile and KeyFile file locations")
}

func TestInvalidClientOptionsWithCreator(t *testing.T) {
	invalidOptions := types.MessageBusConfig{PublishHost: types.HostInfo{
		Host:     "    ",
		Port:     0,
		Protocol: "    ",
	}}

	client, _ := NewMQTTClientWithCreator(invalidOptions, json.Marshal, json.Unmarshal, DefaultClientCreator())

	err := client.Connect()
	// Expecting a timeout error since creating mqtt client now at the beginning of the Connect() function
	_, ok := err.(TimeoutErr)
	if ok {
		err = nil
	}

	require.Error(t, err)
}

func TestClient_Connect(t *testing.T) {
	tests := []struct {
		name         string
		connectToken MockToken
		expectError  bool
		errorType    error
	}{
		{
			"Successful connection",
			SuccessfulMockToken(),
			false,
			nil,
		},
		{
			"Connect timeout with error",
			TimeoutWithErrorMockToken(),
			true,
			TimeoutErr{},
		},
		{
			"Connect timeout without error",
			TimeoutNoErrorMockToken(),
			true,
			TimeoutErr{},
		},
		{
			"Connect error",
			ErrorMockToken(),
			true,
			OperationErr{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			client, _ := NewMQTTClientWithCreator(
				TestMessageBusConfig,
				json.Marshal,
				json.Unmarshal,
				mockClientCreator(test.connectToken, MockToken{}, MockToken{}))

			err := client.Connect()
			if !test.expectError && err != nil {
				t.Errorf("Did not expect error but observed: %s", err.Error())
				return
			}

			if test.expectError {
				require.Error(t, err)
			}

			if test.errorType != nil {

				eet := reflect.TypeOf(test.errorType)
				aet := reflect.TypeOf(err)
				assert.Condition(t, func() (success bool) {
					return aet.AssignableTo(eet)
				}, "Expected error of type %v, but got an error of type %v", eet, aet)
			}
		})
	}
}

func TestClient_Publish(t *testing.T) {
	tests := []struct {
		name         string
		publishToken MockToken
		message      types.MessageEnvelope
		marshaller   MessageMarshaller
		expectError  bool
		errorType    error
	}{
		{
			"Successful publish",
			SuccessfulMockToken(),
			types.MessageEnvelope{},
			json.Marshal,
			false,
			nil,
		},

		{
			"Marshal error",
			SuccessfulMockToken(),
			types.MessageEnvelope{},
			mockMarshallerError,
			true,
			OperationErr{},
		},
		{
			"Publish error",
			ErrorMockToken(),
			types.MessageEnvelope{},
			json.Marshal,
			true,
			OperationErr{},
		},
		{
			"Publish timeout with error",
			TimeoutWithErrorMockToken(),
			types.MessageEnvelope{},
			json.Marshal,
			true,
			TimeoutErr{},
		},
		{
			"Publish timeout without error",
			TimeoutNoErrorMockToken(),
			types.MessageEnvelope{},
			json.Marshal,
			true,
			TimeoutErr{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _ := NewMQTTClientWithCreator(
				TestMessageBusConfig,
				test.marshaller,
				json.Unmarshal,
				mockClientCreator(SuccessfulMockToken(), test.publishToken, MockToken{}))

			err := client.Connect()
			require.NoError(t, err)

			err = client.Publish(test.message, "test-topic")
			if test.expectError {
				require.Error(t, err)
				return // End test for expected error
			} else {
				require.NoError(t, err)
			}

			if test.errorType != nil {
				eet := reflect.TypeOf(test.errorType)
				aet := reflect.TypeOf(err)
				assert.Condition(t, func() (success bool) {
					return aet.AssignableTo(eet)
				}, "Expected error of type %v, but got an error of type %v", eet, aet)
			}
		})
	}
}

func TestClient_Subscribe(t *testing.T) {
	tests := []struct {
		name           string
		subscribeToken MockToken
		topics         []string
		expectError    bool
		errorType      error
	}{
		{
			"Successful subscription",
			SuccessfulMockToken(),
			[]string{"topic"},
			false,
			nil,
		},
		{
			"Successful subscription multiple topics",
			SuccessfulMockToken(),
			[]string{"topic1", "topic2", "topic3"},
			false,
			nil,
		},
		{
			"Subscribe error",
			ErrorMockToken(),
			[]string{"topic1", "topic2"},
			true,
			OperationErr{},
		},
		{
			"Subscribe timeout with error",
			TimeoutWithErrorMockToken(),
			[]string{"topic1", "topic2"},
			true,
			TimeoutErr{},
		},
		{
			"Subscribe timeout without error",
			TimeoutNoErrorMockToken(),
			[]string{"topic1", "topic2"},
			true,
			TimeoutErr{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _ := NewMQTTClientWithCreator(
				TestMessageBusConfig,
				json.Marshal,
				json.Unmarshal,
				mockClientCreator(MockToken{}, MockToken{}, test.subscribeToken))
			topicChannels := make([]types.TopicChannel, 0)
			for _, topic := range test.topics {
				topicChannels = append(topicChannels, types.TopicChannel{
					Topic:    topic,
					Messages: make(chan types.MessageEnvelope),
				})
			}

			err := client.Connect()
			// Expecting a timeout error since creating mqtt client now at the beginning of the Connect() function
			_, ok := err.(TimeoutErr)
			if ok {
				err = nil
			}

			err = client.Subscribe(topicChannels, make(chan error))
			if test.expectError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			if test.errorType != nil {
				eet := reflect.TypeOf(test.errorType)
				aet := reflect.TypeOf(err)
				assert.Condition(t, func() (success bool) {
					return aet.AssignableTo(eet)
				}, "Expected error of type %v, but got an error of type %v", eet, aet)
			}
		})
	}
}

func TestClient_Disconnect(t *testing.T) {
	client, _ := NewMQTTClient(TestMessageBusConfig)

	err := client.Connect()
	// Expecting a timeout error since creating mqtt client now at the beginning of the Connect() function
	_, ok := err.(TimeoutErr)
	if ok {
		err = nil
	}

	err = client.Disconnect()
	require.NoError(t, err, "Disconnect is not expected to return an errors")

	err = client.Disconnect()
	require.NoError(t, err, "Disconnect is not expected to return an error if not connected")
}

func TestSubscriptionMessageHandler(t *testing.T) {
	client, _ := NewMQTTClientWithCreator(
		TestMessageBusConfig,
		json.Marshal,
		json.Unmarshal,
		mockClientCreator(SuccessfulMockToken(), SuccessfulMockToken(), SuccessfulMockToken()))

	err := client.Connect()
	require.NoError(t, err)

	topic1Channel := make(chan types.MessageEnvelope)
	topic2Channel := make(chan types.MessageEnvelope)
	topicChannels := []types.TopicChannel{{
		Topic:    "test1",
		Messages: topic1Channel,
	}, {
		Topic:    "test2",
		Messages: topic2Channel,
	}}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go receiveMessage(wg, topic1Channel, 1)
	go receiveMessage(wg, topic2Channel, 1)
	err = client.Subscribe(topicChannels, make(chan error))
	require.NoError(t, err)
	err = client.Publish(types.MessageEnvelope{
		Checksum:      "123",
		CorrelationID: "456",
		Payload:       []byte("Simple payload"),
		ContentType:   "application/json",
	}, "test1")
	require.NoError(t, err)
	err = client.Publish(types.MessageEnvelope{
		Checksum:      "789",
		CorrelationID: "000",
		Payload:       []byte("Another simple payload"),
		ContentType:   "application/json",
	}, "test2")
	require.NoError(t, err)
	wg.Wait()
}

func TestSubscriptionMessageHandlerError(t *testing.T) {
	client, _ := NewMQTTClientWithCreator(TestMessageBusConfig,
		json.Marshal,
		mockUnmarshalerError,
		mockClientCreator(MockToken{
			waitTimeOut: true,
			err:         nil,
		}, MockToken{
			waitTimeOut: true,
			err:         nil,
		}, MockToken{
			waitTimeOut: true,
			err:         nil,
		}))

	err := client.Connect()
	require.NoError(t, err)

	topicChannels := []types.TopicChannel{{
		Topic:    "test1",
		Messages: make(chan types.MessageEnvelope),
	}, {
		Topic:    "test2",
		Messages: make(chan types.MessageEnvelope),
	}}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	errorChannel := make(chan error)
	go receiveError(wg, errorChannel, 1)
	err = client.Subscribe(topicChannels, errorChannel)
	require.NoError(t, err)
	err = client.Publish(types.MessageEnvelope{
		Checksum:      "123",
		CorrelationID: "456",
		Payload:       []byte("Simple payload"),
		ContentType:   "application/json",
	}, "test1")
	require.NoError(t, err)
	err = client.Publish(types.MessageEnvelope{
		Checksum:      "789",
		CorrelationID: "000",
		Payload:       []byte("Another simple payload"),
		ContentType:   "application/json",
	}, "test2")
	require.NoError(t, err)
	wg.Wait()
}

// mockMarshallerError returns an error when marshaling is attempted.
func mockMarshallerError(interface{}) ([]byte, error) {
	return nil, errors.New("marshal error")
}

// mockUnmarshalerError returns an error when unmarshaling is attempted.
func mockUnmarshalerError([]byte, interface{}) error {
	return errors.New("unmarshal error")
}

// receiveMessage polls the provided channel until the expected number of messages has been received.
func receiveMessage(group *sync.WaitGroup, messageChannel <-chan types.MessageEnvelope, expectedMessages int) {
	for counter := 0; counter < expectedMessages; counter++ {
		<-messageChannel
	}
	group.Done()
}

// receiveError polls the provided channel until the expected number of errors has been received.
func receiveError(group *sync.WaitGroup, errorChannel <-chan error, expectedMessages int) {
	for counter := 0; counter < expectedMessages; counter++ {
		<-errorChannel
	}
	group.Done()
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
