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

package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v3/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	pahoMqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

var OptionalPropertiesNoTls = map[string]string{
	pkg.Username:     "TestUser",
	pkg.Password:     "TestPassword",
	pkg.ClientId:     "TestClientID",
	pkg.Qos:          "1",
	pkg.KeepAlive:    "3",
	pkg.Retained:     "true",
	pkg.CleanSession: "false",
}

var OptionalPropertiesCertCreate = map[string]string{
	pkg.Username:       "TestUser",
	pkg.Password:       "TestPassword",
	pkg.ClientId:       "TestClientID",
	pkg.Qos:            "1",
	pkg.KeepAlive:      "3",
	pkg.Retained:       "true",
	pkg.CleanSession:   "false",
	pkg.CertPEMBlock:   "CertBytes",
	pkg.KeyPEMBlock:    "KeyBytes",
	pkg.ConnectTimeout: "1",
}

var OptionalPropertiesCertLoad = map[string]string{
	pkg.Username:       "TestUser",
	pkg.Password:       "TestPassword",
	pkg.ClientId:       "TestClientID",
	pkg.Qos:            "1",
	pkg.KeepAlive:      "3",
	pkg.Retained:       "true",
	pkg.CleanSession:   "false",
	pkg.CertFile:       "./cert",
	pkg.KeyFile:        "./key",
	pkg.ConnectTimeout: "1",
}

var OptionalPropertiesCaCertCreate = map[string]string{
	pkg.Username:       "TestUser",
	pkg.Password:       "TestPassword",
	pkg.ClientId:       "TestClientID",
	pkg.Qos:            "1",
	pkg.KeepAlive:      "3",
	pkg.Retained:       "true",
	pkg.CleanSession:   "false",
	pkg.CaPEMBlock:     "CertBytes",
	pkg.ConnectTimeout: "1",
}

var OptionalPropertiesCaCertLoad = map[string]string{
	pkg.Username:       "TestUser",
	pkg.Password:       "TestPassword",
	pkg.ClientId:       "TestClientID",
	pkg.Qos:            "1",
	pkg.KeepAlive:      "3",
	pkg.Retained:       "true",
	pkg.CleanSession:   "false",
	pkg.CaFile:         "./cacert",
	pkg.ConnectTimeout: "1",
}

var TcpHostInfo = types.HostInfo{Host: "localhost", Protocol: "tcp", Port: 1883}
var TlsHostInfo = types.HostInfo{Host: "localhost", Protocol: "tls", Port: 8883}
var TcpsHostInfo = types.HostInfo{Host: "localhost", Protocol: "tcps", Port: 8883}
var SslHostInfo = types.HostInfo{Host: "localhost", Protocol: "ssl", Port: 8883}

// TestMessageBusConfig defines a simple configuration used for testing successful options parsing.
var TestMessageBusConfig = types.MessageBusConfig{
	Broker:   TcpsHostInfo,
	Optional: OptionalPropertiesNoTls,
}
var TestMessageBusConfigTlsCreate = types.MessageBusConfig{
	Broker:   TlsHostInfo,
	Optional: OptionalPropertiesCertCreate,
}

var TestMessageBusConfigTlsLoad = types.MessageBusConfig{
	Broker:   TlsHostInfo,
	Optional: OptionalPropertiesCertLoad,
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

	go handler(mc, MockMessage{payload: message.([]byte), topic: topic})
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

func (mt MockToken) Done() <-chan struct{} {
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

func (mc MockMQTTClient) Unsubscribe(topics ...string) pahoMqtt.Token {
	for _, topic := range topics {
		mc.subscriptions[topic] = nil
	}

	return &mc.subscribe
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
	topic   string
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

func (mm MockMessage) Topic() string {
	return mm.topic
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
	invalidOptions := types.MessageBusConfig{Broker: types.HostInfo{
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
		Broker: TlsHostInfo,
	}
	tests := []struct {
		name           string
		optionalConfig map[string]string
	}{
		{name: "Client cert", optionalConfig: map[string]string{
			"CertFile": "./does-not-exist", "KeyFile": "./does-not-exist"}},
		{name: "CA cert", optionalConfig: map[string]string{"CaFile": "./does-not-exist"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options.Optional = test.optionalConfig
			client, _ := NewMQTTClient(options)
			err := client.Connect()
			require.Error(t, err)
		})
	}
}

func TestClientCreatorTLS(t *testing.T) {
	tests := []struct {
		name            string
		hostConfig      types.HostInfo
		optionalConfig  map[string]string
		certCreator     pkg.X509KeyPairCreator
		certLoader      pkg.X509KeyLoader
		caCertCreator   pkg.X509CaCertCreator
		caCertLoader    pkg.X509CaCertLoader
		pemDecoder      pkg.PEMDecoder
		expectError     bool
		expectTLSConfig bool
	}{
		{
			name:       "Create TLS Config from PEM Block (clientcert)",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				pkg.CertPEMBlock:   "CertPEMBlock",
				pkg.KeyPEMBlock:    "KeyPEMBlock",
				pkg.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Create TLS Config from PEM Block (cacert)",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				pkg.CaPEMBlock:     "CaPEMBlock",
				pkg.ConnectTimeout: "1",
			},
			caCertCreator:   mockCaCertCreator(nil),
			pemDecoder:      mockPemDecoder(&pem.Block{}),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Create TCPS Config from PEM Block (clientcert)",
			hostConfig: TcpsHostInfo,
			optionalConfig: map[string]string{
				pkg.CertPEMBlock:   "CertPEMBlock",
				pkg.KeyPEMBlock:    "KeyPEMBlock",
				pkg.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Create TCPS Config from PEM Block (cacert)",
			hostConfig: TcpsHostInfo,
			optionalConfig: map[string]string{
				pkg.CaPEMBlock:     "CaPEMBlock",
				pkg.ConnectTimeout: "1",
			},
			caCertCreator:   mockCaCertCreator(nil),
			pemDecoder:      mockPemDecoder(&pem.Block{}),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Create SSL Config from PEM Block (clientcert)",
			hostConfig: SslHostInfo,
			optionalConfig: map[string]string{
				pkg.CertPEMBlock:   "CertPEMBlock",
				pkg.KeyPEMBlock:    "KeyPEMBlock",
				pkg.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Create SSL Config from PEM Block (cacert)",
			hostConfig: SslHostInfo,
			optionalConfig: map[string]string{
				pkg.CaPEMBlock:     "CaPEMBlock",
				pkg.ConnectTimeout: "1",
			},
			caCertCreator:   mockCaCertCreator(nil),
			pemDecoder:      mockPemDecoder(&pem.Block{}),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Skip TLS Config from PEM Block for non-supported TLS protocols",
			hostConfig: TcpHostInfo,
			optionalConfig: map[string]string{
				pkg.CertPEMBlock:   "CertPEMBlock",
				pkg.KeyPEMBlock:    "KeyPEMBlock",
				pkg.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: false,
		},
		{
			name:       "Fail Create TLS Config from PEM File (clientcert)",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				pkg.CertPEMBlock:   "CertPEMBlock",
				pkg.KeyPEMBlock:    "KeyPEMBlock",
				pkg.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(errors.New("test error")),
			certLoader:      mockCertLoader(nil),
			expectError:     true,
			expectTLSConfig: false,
		},
		{
			name:       "Fail Create TLS Config from PEM File (cacert)",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				pkg.CaFile:         "./does-not-exist",
				pkg.ConnectTimeout: "1",
			},
			caCertLoader:    mockCaCertLoader(errors.New("test error")),
			pemDecoder:      mockPemDecoder(&pem.Block{}),
			expectError:     true,
			expectTLSConfig: false,
		},
		{
			name:       "Load TLS Config from Cert File (clientcert)",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				pkg.CertFile:       "./cert",
				pkg.KeyFile:        "./key",
				pkg.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Load TLS Config from Cert File (cacert)",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				pkg.CaFile:         "./cacert",
				pkg.ConnectTimeout: "1",
			},
			caCertCreator:   mockCaCertCreator(nil),
			caCertLoader:    mockCaCertLoader(nil),
			pemDecoder:      mockPemDecoder(&pem.Block{}),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Load TCPS Config from Cert File (clientcert)",
			hostConfig: TcpsHostInfo,
			optionalConfig: map[string]string{
				pkg.CertFile:       "./cert",
				pkg.KeyFile:        "./key",
				pkg.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Load TCPS Config from Cert File (cacert)",
			hostConfig: TcpsHostInfo,
			optionalConfig: map[string]string{
				pkg.CaFile:         "./cacert",
				pkg.ConnectTimeout: "1",
			},
			caCertCreator:   mockCaCertCreator(nil),
			caCertLoader:    mockCaCertLoader(nil),
			pemDecoder:      mockPemDecoder(&pem.Block{}),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Load SSL Config from Cert File (clientcert)",
			hostConfig: SslHostInfo,
			optionalConfig: map[string]string{
				pkg.CertFile:       "./cert",
				pkg.KeyFile:        "./key",
				pkg.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Load SSL Config from Cert File (cacert)",
			hostConfig: SslHostInfo,
			optionalConfig: map[string]string{
				pkg.CaFile:         "./cacert",
				pkg.ConnectTimeout: "1",
			},
			caCertCreator:   mockCaCertCreator(nil),
			caCertLoader:    mockCaCertLoader(nil),
			pemDecoder:      mockPemDecoder(&pem.Block{}),
			expectError:     false,
			expectTLSConfig: true,
		},
		{
			name:       "Skip Load TLS Config from Cert File for un-supported protocols",
			hostConfig: TcpHostInfo,
			optionalConfig: map[string]string{
				pkg.CertFile:       "./cert",
				pkg.KeyFile:        "./key",
				pkg.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(nil),
			expectError:     false,
			expectTLSConfig: false,
		},
		{
			name:       "Fail Load TLS Config from Cert File (clientcert)",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				pkg.CertFile:       "./cert",
				pkg.KeyFile:        "./key",
				pkg.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(errors.New("test error")),
			expectError:     true,
			expectTLSConfig: false,
		},
		{
			name:       "Fail Load TLS Config from Cert File (cacert)",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				pkg.CaFile:         "./cacert",
				pkg.ConnectTimeout: "1",
			},
			caCertLoader:    mockCaCertLoader(errors.New("test error")),
			expectError:     true,
			expectTLSConfig: false,
		},
		{
			name:       "Fail Load TLS Config For Invalid Options",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				pkg.Qos:            "abc",
				pkg.ConnectTimeout: "1",
			},
			certCreator:     mockCertCreator(nil),
			certLoader:      mockCertLoader(errors.New("test error")),
			expectError:     true,
			expectTLSConfig: false,
		},
		{
			name:       "Fail Load TLS Config For Empty PEM Block (cacert)",
			hostConfig: TlsHostInfo,
			optionalConfig: map[string]string{
				pkg.CaFile:         "./cacert",
				pkg.ConnectTimeout: "1",
			},
			caCertLoader:    mockCaCertLoader(nil),
			pemDecoder:      mockPemDecoder(nil),
			expectError:     true,
			expectTLSConfig: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _ := NewMQTTClientWithCreator(
				types.MessageBusConfig{
					Broker:   test.hostConfig,
					Optional: test.optionalConfig,
				},
				json.Marshal,
				json.Unmarshal,
				ClientCreatorWithCertLoader(test.certCreator, test.certLoader, test.caCertCreator, test.caCertLoader,
					test.pemDecoder))

			err := client.Connect()

			// Expecting a connect error since creating mqtt client now at the beginning of the Connect() function
			if err != nil &&
				(strings.Contains(err.Error(), "connect: connection refused") || // Linux
					strings.Contains(err.Error(), "Unable to connect")) { // Windows
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
	tests := []struct {
		name           string
		optionalConfig map[string]string
	}{
		{name: "Client cert", optionalConfig: OptionalPropertiesCertLoad},
		{name: "CA cert", optionalConfig: OptionalPropertiesCaCertLoad},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			TestMessageBusConfigTlsLoad.Optional = test.optionalConfig
			client, _ := NewMQTTClientWithCreator(
				TestMessageBusConfigTlsLoad,
				json.Marshal,
				json.Unmarshal,
				ClientCreatorWithCertLoader(mockCertCreator(nil), mockCertLoader(nil),
					mockCaCertCreator(nil), mockCaCertLoader(nil), mockPemDecoder(&pem.Block{})))

			// Expect Connect to return an error since no broker available
			err := client.Connect()
			require.Error(t, err)

			clientOptions := client.mqttClient.OptionsReader()
			tlsConfig := clientOptions.TLSConfig()
			assert.NotNil(t, tlsConfig, "Failed to configure TLS for underlying client")
		})
	}
}

func TestClientCreatorTlsLoadError(t *testing.T) {
	tests := []struct {
		name           string
		optionalConfig map[string]string
	}{
		{name: "Client cert", optionalConfig: OptionalPropertiesCertLoad},
		{name: "CA cert", optionalConfig: OptionalPropertiesCaCertLoad},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			TestMessageBusConfigTlsLoad.Optional = test.optionalConfig
			client, _ := NewMQTTClientWithCreator(
				TestMessageBusConfigTlsLoad,
				json.Marshal,
				json.Unmarshal,
				ClientCreatorWithCertLoader(mockCertCreator(nil), mockCertLoader(errors.New("test error")),
					mockCaCertCreator(nil), mockCaCertLoader(errors.New("test error")), mockPemDecoder(&pem.Block{})))

			err := client.Connect()
			// Expecting a timeout error since creating mqtt client now at the beginning of the Connect() function
			_, ok := err.(TimeoutErr)
			if ok {
				err = nil
			}

			assert.Error(t, err, "Expected error for invalid CertFile and KeyFile file locations")
		})
	}
}

func TestClientCreatorTlsCreator(t *testing.T) {
	tests := []struct {
		name           string
		optionalConfig map[string]string
	}{
		{name: "Client cert", optionalConfig: OptionalPropertiesCertCreate},
		{name: "CA cert", optionalConfig: OptionalPropertiesCaCertCreate},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			TestMessageBusConfigTlsCreate.Optional = test.optionalConfig
			client, _ := NewMQTTClientWithCreator(
				TestMessageBusConfigTlsCreate,
				json.Marshal,
				json.Unmarshal,
				ClientCreatorWithCertLoader(mockCertCreator(nil), mockCertLoader(nil),
					mockCaCertCreator(nil), mockCaCertLoader(nil), mockPemDecoder(&pem.Block{})))

			// Expect Connect to return an error since no broker available
			err := client.Connect()
			require.Error(t, err)

			clientOptions := client.mqttClient.OptionsReader()
			tlsConfig := clientOptions.TLSConfig()
			assert.NotNil(t, tlsConfig, "Failed to configure TLS for underlying client")
		})
	}
}

func TestClientCreatorTlsCreatorError(t *testing.T) {
	tests := []struct {
		name           string
		optionalConfig map[string]string
	}{
		{name: "Client cert", optionalConfig: OptionalPropertiesCertCreate},
		{name: "CA cert", optionalConfig: OptionalPropertiesCaCertCreate},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			TestMessageBusConfigTlsCreate.Optional = test.optionalConfig
			client, _ := NewMQTTClientWithCreator(
				TestMessageBusConfigTlsCreate,
				json.Marshal,
				json.Unmarshal,
				ClientCreatorWithCertLoader(mockCertCreator(errors.New("test error")), mockCertLoader(nil),
					mockCaCertCreator(errors.New("test error")), mockCaCertLoader(nil), mockPemDecoder(&pem.Block{})))

			err := client.Connect()
			// Expecting a timeout error since creating mqtt client now at the beginning of the Connect() function
			_, ok := err.(TimeoutErr)
			if ok {
				err = nil
			}

			assert.Error(t, err, "Expected error for invalid CertFile and KeyFile file locations")
		})
	}
}

func TestInvalidClientOptionsWithCreator(t *testing.T) {
	invalidOptions := types.MessageBusConfig{Broker: types.HostInfo{
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
			require.NoError(t, err)

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

func TestClient_Unsubscribe(t *testing.T) {
	target, err := NewMQTTClientWithCreator(
		TestMessageBusConfig,
		json.Marshal,
		json.Unmarshal,
		mockClientCreator(SuccessfulMockToken(), SuccessfulMockToken(), SuccessfulMockToken()))
	require.NoError(t, err)

	err = target.Connect()
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

	_, exists := target.existingSubscriptions["test1"]
	require.True(t, exists)
	_, exists = target.existingSubscriptions["test2"]
	require.True(t, exists)
	_, exists = target.existingSubscriptions["test3"]
	require.True(t, exists)

	err = target.Unsubscribe("test1")
	require.NoError(t, err)

	_, exists = target.existingSubscriptions["test1"]
	require.False(t, exists)
	_, exists = target.existingSubscriptions["test2"]
	require.True(t, exists)
	_, exists = target.existingSubscriptions["test3"]
	require.True(t, exists)

	err = target.Unsubscribe("test1", "test2", "test3")
	require.NoError(t, err)

	_, exists = target.existingSubscriptions["test1"]
	require.False(t, exists)
	_, exists = target.existingSubscriptions["test2"]
	require.False(t, exists)
	_, exists = target.existingSubscriptions["test3"]
	require.False(t, exists)
}

func TestClient_Disconnect(t *testing.T) {
	client, _ := NewMQTTClient(TestMessageBusConfig)

	err := client.Connect()
	// Expecting a timeout error since creating mqtt client now at the beginning of the Connect() function
	_, ok := err.(TimeoutErr)
	if ok {
		err = nil
	}
	assert.Error(t, err)

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
	go receiveMessage(t, wg, topicChannels[0], 1)
	go receiveMessage(t, wg, topicChannels[1], 1)
	err = client.Subscribe(topicChannels, make(chan error))
	require.NoError(t, err)
	err = client.Publish(types.MessageEnvelope{
		CorrelationID: "456",
		Payload:       []byte("Simple payload"),
		ContentType:   "application/json",
	}, "test1")
	require.NoError(t, err)
	err = client.Publish(types.MessageEnvelope{
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
		CorrelationID: "456",
		Payload:       []byte("Simple payload"),
		ContentType:   "application/json",
	}, "test1")
	require.NoError(t, err)
	err = client.Publish(types.MessageEnvelope{
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
func receiveMessage(t *testing.T, group *sync.WaitGroup, topicChannel types.TopicChannel, expectedMessages int) {
	for counter := 0; counter < expectedMessages; counter++ {
		message := <-topicChannel.Messages
		// Unit test is using simple topic, i.e. without wild card
		assert.Equal(t, topicChannel.Topic, message.ReceivedTopic, "ReceivedTopic not as expected")
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

func TestCreateClientOptions(t *testing.T) {
	config := MQTTClientConfig{
		BrokerURL: "",
		MQTTClientOptions: MQTTClientOptions{
			ClientId:       "Test",
			Qos:            2,
			KeepAlive:      50,
			Retained:       true,
			AutoReconnect:  true,
			CleanSession:   false,
			ConnectTimeout: 60,
			TlsConfigurationOptions: pkg.TlsConfigurationOptions{
				SkipCertVerify: false,
				CertFile:       "",
				KeyFile:        "",
				KeyPEMBlock:    "",
				CertPEMBlock:   "",
			},
		},
	}

	options, err := createClientOptions(config, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	assert.Equal(t, config.ClientId, options.ClientID)
	assert.Equal(t, byte(config.Qos), options.WillQos)
	assert.Equal(t, config.Retained, options.WillRetained)
	assert.Equal(t, int64(config.KeepAlive), options.KeepAlive)
	assert.Equal(t, config.AutoReconnect, options.AutoReconnect)
	assert.Equal(t, time.Duration(config.ConnectTimeout)*time.Second, options.ConnectTimeout)
}
