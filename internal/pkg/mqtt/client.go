/********************************************************************************
 *  Copyright 2019 Dell Inc.
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
	"fmt"
	"net/url"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var TlsSchemes = []string{"tcps", "ssl", "tls"}

// ClientCreator defines the function signature for creating an MQTT client.
type ClientCreator func(config types.MessageBusConfig) (mqtt.Client, error)

// X509KeyPairCreator defines the function signature for creating a tls.Certificate based on PEM encoding.
type X509KeyPairCreator func(certPEMBlock []byte, keyPEMBlock []byte) (tls.Certificate, error)

// X509KeyLoader defines a function signature for loading a tls.Certificate from cert and key files.
type X509KeyLoader func(certFile string, keyFile string) (tls.Certificate, error)

// MessageMarshaler defines the function signature for marshaling structs into []byte.
type MessageMarshaler func(v interface{}) ([]byte, error)

// MessageUnmarshaler defines the function signature for unmarshaling []byte into structs.
type MessageUnmarshaler func(data []byte, v interface{}) error

// Client facilitates communication to an MQTT server and provides functionality needed to send and receive MQTT
// messages.
type Client struct {
	wrappedClient mqtt.Client
	marshaler     MessageMarshaler
	unmarshaler   MessageUnmarshaler
}

// NewMQTTClient constructs a new MQTT client based on the options provided.
func NewMQTTClient(options types.MessageBusConfig) (Client, error) {
	mqttClient, err := DefaultClientCreator()(options)
	if err != nil {
		return Client{}, err
	}

	return Client{
		wrappedClient: mqttClient,
		marshaler:     json.Marshal,
		unmarshaler:   json.Unmarshal,
	}, nil
}

// NewMQTTClientWithCreator constructs a new MQTT client based on the options and ClientCreator provided.
func NewMQTTClientWithCreator(
	options types.MessageBusConfig,
	marshaler MessageMarshaler,
	unmarshaler MessageUnmarshaler,
	creator ClientCreator) (Client, error) {

	wrappedClient, err := creator(options)
	if err != nil {
		return Client{}, err
	}

	return Client{
		wrappedClient: wrappedClient,
		marshaler:     marshaler,
		unmarshaler:   unmarshaler,
	}, nil
}

// Connect establishes a connection to a MQTT server.
// This must be called before any other functionality provided by the Client.
func (mc Client) Connect() error {
	optionsReader := mc.wrappedClient.OptionsReader()

	return getTokenError(
		mc.wrappedClient.Connect(),
		optionsReader.ConnectTimeout(),
		ConnectOperation,
		"Unable to connect")
}

// Publish sends a message to the connected MQTT server.
func (mc Client) Publish(message types.MessageEnvelope, topic string) error {
	marshaledMessage, err := mc.marshaler(message)
	if err != nil {
		return NewOperationErr(PublishOperation, err.Error())
	}

	optionsReader := mc.wrappedClient.OptionsReader()
	return getTokenError(
		mc.wrappedClient.Publish(
			topic,
			optionsReader.WillQos(),
			optionsReader.WillRetained(),
			marshaledMessage),
		optionsReader.ConnectTimeout(),
		PublishOperation,
		"Unable to publish message")

}

// Subscribe creates a subscription for the specified topics.
func (mc Client) Subscribe(topics []types.TopicChannel, messageErrors chan error) error {
	optionsReader := mc.wrappedClient.OptionsReader()

	for _, topic := range topics {
		err := getTokenError(
			mc.wrappedClient.Subscribe(
				topic.Topic,
				optionsReader.WillQos(),
				newMessageHandler(mc.unmarshaler, topic.Messages, messageErrors)),
			optionsReader.ConnectTimeout(),
			SubscribeOperation,
			"Failed to create subscription")
		if err != nil {
			return err
		}
	}

	return nil
}

// Disconnect closes the connection to the connected MQTT server.
func (mc Client) Disconnect() error {

	// Specify a wait time equal to the write timeout so that we allow other any queued processing to complete before
	// disconnecting.
	optionsReader := mc.wrappedClient.OptionsReader()
	mc.wrappedClient.Disconnect(uint(optionsReader.ConnectTimeout() * time.Millisecond))

	return nil
}

// DefaultClientCreator returns a default function for creating MQTT clients.
func DefaultClientCreator() ClientCreator {
	return func(options types.MessageBusConfig) (mqtt.Client, error) {
		clientConfiguration, err := CreateMQTTClientConfiguration(options)
		if err != nil {
			return nil, err
		}

		clientOptions, err := createClientOptions(clientConfiguration, tls.X509KeyPair, tls.LoadX509KeyPair)
		if err != nil {
			return nil, err
		}

		return mqtt.NewClient(clientOptions), nil
	}
}

// ClientCreatorWithCertLoader creates a ClientCreator which leverages the specified cert creator and loader when
// creating an MQTT client.
func ClientCreatorWithCertLoader(certCreator X509KeyPairCreator, certLoader X509KeyLoader) ClientCreator {
	return func(options types.MessageBusConfig) (mqtt.Client, error) {
		clientConfiguration, err := CreateMQTTClientConfiguration(options)
		if err != nil {
			return nil, err
		}

		clientOptions, err := createClientOptions(clientConfiguration, certCreator, certLoader)
		if err != nil {
			return nil, err
		}

		return mqtt.NewClient(clientOptions), nil
	}
}

// newMessageHandler creates a function which meets the criteria for a MessageHandler and propagates the received
// messages to the proper channel.
func newMessageHandler(
	unmarshaler MessageUnmarshaler,
	messageChannel chan<- types.MessageEnvelope,
	errorChannel chan<- error) mqtt.MessageHandler {

	return func(client mqtt.Client, message mqtt.Message) {
		var messageEnvelope types.MessageEnvelope
		payload := message.Payload()
		err := unmarshaler(payload, &messageEnvelope)
		if err != nil {
			errorChannel <- err
		}

		messageChannel <- messageEnvelope
	}
}

// getTokenError determines if a Token is in an errored state and if so returns the proper error message. Otherwise,
// nil.
//
// NOTE the paho.mqtt.golang's recommended way for handling errors do not cover all cases. During manual verification
// with an MQTT server, it was observed that the Token.Error() was sometimes nil even when a token.WaitTimeout(...)
// returned false(indicating the operation has timed-out). Therefore, there are some additional checks that need to
// take place to ensure the error message is returned if it is present. One example scenario, if you attempt to connect
// without providing a ClientID.
func getTokenError(token mqtt.Token, timeout time.Duration, operation string, defaultTimeoutMessage string) error {
	hasTimedOut := !token.WaitTimeout(timeout)

	if hasTimedOut && token.Error() != nil {
		return NewTimeoutError(operation, token.Error().Error())
	}

	if hasTimedOut && token.Error() == nil {
		return NewTimeoutError(operation, defaultTimeoutMessage)
	}

	if token.Error() != nil {
		return NewOperationErr(operation, token.Error().Error())
	}

	return nil
}

// createClientOptions constructs mqtt.Client options from an MQTTClientConfig.
func createClientOptions(
	clientConfiguration MQTTClientConfig,
	certCreator X509KeyPairCreator,
	certLoader X509KeyLoader) (*mqtt.ClientOptions, error) {

	clientOptions := mqtt.NewClientOptions()
	clientOptions.AddBroker(clientConfiguration.BrokerURL)
	clientOptions.SetUsername(clientConfiguration.Username)
	clientOptions.SetPassword(clientConfiguration.Password)
	clientOptions.SetClientID(clientConfiguration.ClientId)
	clientOptions.SetWill(clientConfiguration.Topic, clientConfiguration.ConnectionPayload, byte(clientConfiguration.Qos), clientConfiguration.Retained)
	clientOptions.SetKeepAlive(time.Duration(clientConfiguration.KeepAlive) * time.Second)
	clientOptions.SetAutoReconnect(clientConfiguration.AutoReconnect)
	tlsConfiguration, err := generateTLSForClientClientOptions(clientConfiguration, certCreator, certLoader)
	if err != nil {
		return clientOptions, err
	}

	clientOptions.SetTLSConfig(tlsConfiguration)

	return clientOptions, nil
}

// generateTLSForClientClientOptions creates a tls.Config which can be used when configuring the underlying MQTT client.
// If TLS is not needed then nil will be returned which can be used to signal no TLS is needed to the MQTT client.
func generateTLSForClientClientOptions(
	clientConfiguration MQTTClientConfig,
	certCreator X509KeyPairCreator,
	certLoader X509KeyLoader) (*tls.Config, error) {

	// Nothing to do if the CertFile and KeyFile OR CertPEMBlock and KeyPEMBlock  properties are not provided.
	if len(clientConfiguration.CertFile) <= 0 && len(clientConfiguration.KeyFile) <= 0 &&
		len(clientConfiguration.CertPEMBlock) <= 0 && len(clientConfiguration.KeyPEMBlock) <= 0 {
		return nil, nil
	}

	brokerURL, err := url.Parse(clientConfiguration.BrokerURL)
	if err != nil {
		return nil, pkg.NewBrokerURLErr(fmt.Sprintf("Failed to parse broker: %v", err))
	}

	for _, scheme := range TlsSchemes {
		if brokerURL.Scheme != scheme {
			continue
		}

		cert, err := generateCertificate(clientConfiguration, certCreator, certLoader)
		if err != nil {
			return nil, err
		}

		tlsConfig := &tls.Config{
			ClientCAs:          nil,
			InsecureSkipVerify: clientConfiguration.SkipCertVerify,
			Certificates:       []tls.Certificate{cert},
		}

		return tlsConfig, nil
	}

	// The scheme being used either does not require TLS or is not supported with this configuration setup.
	return nil, nil
}

// generateCertificate creates a x509 certificate by either loading it from an existing cert and key files, or creates
// a cert and key from the provided PEM bytes.
func generateCertificate(
	clientConfiguration MQTTClientConfig,
	certCreator X509KeyPairCreator,
	certLoader X509KeyLoader) (tls.Certificate, error) {

	var cert tls.Certificate
	var err error

	if clientConfiguration.KeyPEMBlock != "" && clientConfiguration.CertPEMBlock != "" {
		cert, err = certCreator([]byte(clientConfiguration.CertPEMBlock), []byte(clientConfiguration.KeyPEMBlock))
	} else {
		cert, err = certLoader(clientConfiguration.CertFile, clientConfiguration.KeyFile)
	}

	if err != nil {
		return cert, pkg.NewCertificateErr(fmt.Sprintf("Failed loading x509 data: %v", err))
	}

	return cert, nil
}
