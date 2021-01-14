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

package streams

import (
	"crypto/tls"
	"fmt"

	"github.com/edgexfoundry/go-mod-messaging/v2/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

const (
	// GoModMessagingNamespace go-mod-messaging(gmm) namespace used for grouping and isolating topics(streams). This is
	// prepended to topics when either subscribing or publishing to a stream.
	GoModMessagingNamespace = "gmm"

	// GoModMessagingNamespaceFormat formatted string which allows for consistent construction of entities within Redis
	// that require a namespace.
	GoModMessagingNamespaceFormat = GoModMessagingNamespace + ":%s"
)

// Client MessageClient implementation which provides functionality for sending and receiving messages using
// RedisStreams.
type Client struct {
	// Client used for functionality related to reading messages
	subscribeClient RedisClient

	// Client used for functionality related to sending messages
	publishClient RedisClient
}

// NewClient creates a new Client based on the provided configuration.
func NewClient(messageBusConfig types.MessageBusConfig) (Client, error) {
	return NewClientWithCreator(messageBusConfig, NewGoRedisClientWrapper, tls.X509KeyPair, tls.LoadX509KeyPair)
}

// NewClientWithCreator creates a new Client based on the provided configuration while allowing more control on the
// creation of the underlying entities such as certs, keys, and Redis clients
func NewClientWithCreator(
	messageBusConfig types.MessageBusConfig,
	creator RedisClientCreator,
	pairCreator pkg.X509KeyPairCreator,
	keyLoader pkg.X509KeyLoader) (Client, error) {

	// Parse Optional configuration properties
	optionalClientConfiguration, err := NewClientConfiguration(messageBusConfig)
	if err != nil {
		return Client{}, err
	}

	// Parse TLS configuration properties
	tlsConfigurationOptions := pkg.TlsConfigurationOptions{}
	err = pkg.Load(messageBusConfig.Optional, &tlsConfigurationOptions)
	if err != nil {
		return Client{}, err
	}

	var publishClient, subscribeClient RedisClient

	// Create underlying client to use when publishing
	if !messageBusConfig.PublishHost.IsHostInfoEmpty() {
		publishClient, err = createRedisClient(
			messageBusConfig.PublishHost.GetHostURL(),
			optionalClientConfiguration,
			tlsConfigurationOptions,
			creator,
			pairCreator,
			keyLoader)

		if err != nil {
			return Client{}, err
		}
	}

	// Create underlying client to use when subscribing
	if !messageBusConfig.SubscribeHost.IsHostInfoEmpty() {
		subscribeClient, err = createRedisClient(
			messageBusConfig.SubscribeHost.GetHostURL(),
			optionalClientConfiguration,
			tlsConfigurationOptions,
			creator,
			pairCreator,
			keyLoader)

		if err != nil {
			return Client{}, err
		}
	}

	return Client{
		subscribeClient: subscribeClient,
		publishClient:   publishClient,
	}, nil
}

// Connect noop as preemptive connections are not needed.
func (c Client) Connect() error {
	// No need to connect, connection pooling is handled by the underlying client.
	return nil
}

// Publish sends the provided message to appropriate Redis stream.
func (c Client) Publish(message types.MessageEnvelope, topic string) error {
	if c.publishClient == nil {
		return pkg.NewMissingConfigurationErr("PublishHostInfo", "Unable to create a connection for publishing")
	}

	if topic == "" {
		// Empty streams are not allowed for Redis
		return pkg.NewInvalidTopicErr("", "Unable to publish to the invalid topic")
	}
	values := map[string]interface{}{
		"CorrelationID": message.CorrelationID,
		"Payload":       message.Payload,
		"Checksum":      message.Checksum,
		"ContentType":   message.ContentType,
	}

	return c.publishClient.AddToStream(fmt.Sprintf(GoModMessagingNamespaceFormat, topic), values)
}

// Subscribe creates background processes which reads messages from the appropriate Redis stream and sends to the
// provided channels
func (c Client) Subscribe(topics []types.TopicChannel, messageErrors chan error) error {
	if c.subscribeClient == nil {
		return pkg.NewMissingConfigurationErr("SubscribeHostInfo", "Unable to create a connection for subscribing")
	}

	for _, topic := range topics {
		stream := topic.Topic
		messageChannel := topic.Messages

		go func() {
			for {
				messages, err := c.subscribeClient.ReadFromStream(fmt.Sprintf(GoModMessagingNamespaceFormat, stream))
				if err != nil {
					messageErrors <- err
					continue
				}

				for _, message := range messages {
					messageChannel <- message
				}
			}
		}()
	}

	return nil

}

// Disconnect closes connections to the Redis server.
func (c Client) Disconnect() error {
	var disconnectErrors []string
	if c.publishClient != nil {
		err := c.publishClient.Close()
		if err != nil {
			disconnectErrors = append(disconnectErrors, fmt.Sprintf("Unable to disconnect publish client: %v", err))
		}
	}

	if c.subscribeClient != nil {
		err := c.subscribeClient.Close()
		if err != nil {
			disconnectErrors = append(disconnectErrors, fmt.Sprintf("Unable to disconnect subscribe client: %v", err))
		}

	}

	if len(disconnectErrors) > 0 {
		return NewDisconnectErr(disconnectErrors)
	}

	return nil
}

// createRedisClient helper function for creating RedisClient implementations.
func createRedisClient(
	redisServerURL string,
	optionalClientConfiguration OptionalClientConfiguration,
	tlsConfigurationOptions pkg.TlsConfigurationOptions,
	creator RedisClientCreator,
	pairCreator pkg.X509KeyPairCreator,
	keyLoader pkg.X509KeyLoader) (RedisClient, error) {

	tlsConfig, err := pkg.GenerateTLSForClientClientOptions(
		redisServerURL,
		tlsConfigurationOptions,
		pairCreator,
		keyLoader)

	if err != nil {
		return nil, err
	}

	return creator(redisServerURL, optionalClientConfiguration.Password, tlsConfig)
}
