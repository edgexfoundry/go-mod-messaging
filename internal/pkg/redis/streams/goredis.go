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

// streams package contains a RedisClient which leverages go-redis to interact with a Redis server.
package streams

import (
	"crypto/tls"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	goRedis "github.com/go-redis/redis/v7"
)

// NewGoRedisClientWrapper creates a RedisClient implementation which uses a 'go-redis' Client to achieve the necessary
// functionality.
func NewGoRedisClientWrapper(redisServerURL string, password string, tlsConfig *tls.Config) (RedisClient, error) {
	options, err := goRedis.ParseURL(redisServerURL)
	if err != nil {
		return nil, err
	}

	options.Password = password
	options.TLSConfig = tlsConfig

	return &goRedisWrapper{wrappedClient: goRedis.NewClient(options)}, nil
}

// goRedisWrapper implements RedisClient and uses a underlying 'go-redis' client to communicate with a Redis server.
//
// This functionality was abstracted out from Client so that unit testing can be done easily. The functionality provided
// by this struct can be complex to test and has been tested in the integration test.
type goRedisWrapper struct {
	wrappedClient *goRedis.Client
}

// AddToStream sends the provided message to a stream.
func (g *goRedisWrapper) AddToStream(stream string, values map[string]interface{}) error {
	xAddArgs := &goRedis.XAddArgs{
		Stream:       stream,
		MaxLen:       0,
		MaxLenApprox: 0,
		ID:           "*",
		Values:       values,
	}

	_, err := g.wrappedClient.XAdd(xAddArgs).Result()
	if err != nil {
		return err
	}

	return nil
}

// ReadFromStream retrieves the next message(s) from the specified topic. This operation blocks indefinitely until a
// message received for the topic.
func (g *goRedisWrapper) ReadFromStream(stream string) ([]types.MessageEnvelope, error) {
	readArgs := &goRedis.XReadArgs{
		Streams: []string{stream, LatestStreamMessage},
		Count:   0,
		Block:   0,
	}

	readStreams, err := g.wrappedClient.XRead(readArgs).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]types.MessageEnvelope, 0)
	for _, stream := range readStreams {
		for _, message := range stream.Messages {
			messages = append(messages, types.MessageEnvelope{
				Checksum:      message.Values["Checksum"].(string),
				CorrelationID: message.Values["CorrelationID"].(string),
				Payload:       []byte(message.Values["Payload"].(string)),
				ContentType:   message.Values["ContentType"].(string),
			})
		}
	}

	return messages, nil
}

// Close closes the underlying 'go-redis' client.
func (g *goRedisWrapper) Close() error {
	return g.wrappedClient.Close()
}
