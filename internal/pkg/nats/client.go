//
// Copyright (c) 2022 One Track Consulting
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

//go:build include_nats_messaging

package nats

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-messaging/v2/internal/pkg/nats/interfaces"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/nats-io/nats.go"
)

func newConnection(cc ClientConfig) (interfaces.Connection, error) {
	co, err := cc.ConnectOpt()

	if err != nil {
		return nil, err
	}

	nc, err := nats.Connect(cc.BrokerURL, co...)

	if err != nil {
		return nil, err
	}

	return nc, nil
}

// NewClient initializes creates a new client using NATS core
func NewClient(cfg types.MessageBusConfig) (*Client, error) {
	return NewClientWithConnectionFactory(cfg, newConnection)
}

// NewClientWithConnectionFactory creates a new client that uses the specified function to establish pub/sub connections
func NewClientWithConnectionFactory(cfg types.MessageBusConfig, connectionFactory ConnectNats) (*Client, error) {
	if connectionFactory == nil {
		return nil, fmt.Errorf("connectionFactory is required")
	}

	var m interfaces.MarshallerUnmarshaller = &natsMarshaller{} // default

	cc, err := CreateClientConfiguration(cfg)

	if err != nil {
		return nil, err
	}

	switch strings.ToLower(cc.Format) {
	case "json":
		m = &jsonMarshaller{}
	default:
		m = &natsMarshaller{}
	}

	return &Client{config: cc, connect: connectionFactory, m: m}, nil
}

// Client provides NATS MessageBus implementations per the underlying connection
type Client struct {
	connect    ConnectNats
	connection interfaces.Connection
	m          interfaces.MarshallerUnmarshaller
	config     ClientConfig
}

// Connect establishes the connections to publish and subscribe hosts
func (c *Client) Connect() error {
	if c.connection != nil {
		return fmt.Errorf("already connected to NATS")
	}

	if c.connect == nil {
		return fmt.Errorf("connection function not specified")
	}

	conn, err := c.connect(c.config)

	if err != nil {
		return err
	}

	c.connection = conn

	return nil
}

// Publish publishes EdgeX messages to NATS
func (c *Client) Publish(message types.MessageEnvelope, topic string) error {
	if c.connection == nil {
		return fmt.Errorf("cannot publish with disconnected client")
	}

	if topic == "" {
		return fmt.Errorf("cannot publish to empty topic")
	}

	msg, err := c.m.Marshal(message, topic)

	if err != nil {
		return err
	}

	return c.connection.PublishMsg(msg)
}

// Subscribe establishes NATS subscriptions for the given topics
func (c *Client) Subscribe(topics []types.TopicChannel, messageErrors chan error) error {
	if c.connection == nil {
		return fmt.Errorf("cannot subscribe with disconnected client")
	}

	for _, tc := range topics {
		s := TopicToSubject(tc.Topic)

		_, err := c.connection.QueueSubscribe(s, c.config.QueueGroup, func(msg *nats.Msg) {
			env := types.MessageEnvelope{}
			err := c.m.Unmarshal(msg, &env)
			if err != nil {
				messageErrors <- err
			} else {
				tc.Messages <- env
			}
		})

		if err != nil {
			return err
		}
	}
	return nil
}

// Disconnect drains open subscriptions before closing
func (c *Client) Disconnect() error {
	if c.connection == nil {
		return nil
	}
	return c.connection.Drain()
}
