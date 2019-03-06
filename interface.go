//
// Copyright (c) 2019 Intel Corporation
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

package messaging

// MessageClient is the messaging interface for publisher-subsriber pattern
type MessageClient interface {
	// Connect to messaging host specified in MessageBusconfig
	// returns error if not able to connect
	Connect() error

	// Publish is to send message to the message bus
	// the message contains data payload to send to the message queue
	Publish(message MessageEnvelope, topic string) error

	// Subscribe is to receive messages from topic channels
	// the topic channel contains subscribed message channel and topic to associate with it
	// the channel is used for multiple threads of subscribers for 1 publisher (1-to-many)
	// the messageErrors channel returns the message errors from the caller
	// since subscriber works in asynchrous fashion
	// the function returns error for any subscribe error
	Subscribe(topics []TopicChannel, messageErrors chan error) error

	// Disconnect is to close all connections on the message bus
	Disconnect() error
}

// MessageEnvelope is the data structure for publisher
type MessageEnvelope struct {
	// CorrelationID is an object id to identify the envelop
	CorrelationID string
	// Payload can be JSON marshalled into bytes
	Payload []byte
}

// TopicChannel is the data structure for subscriber
type TopicChannel struct {
	// Topic for subscriber to filter on if any
	Topic string
	// Messages is the returned message channel for the subscriber
	Messages chan interface{}
}

// MessageBusConfig defines the messaging information need to connect to the message bus
// in a publish-subscribe pattern
type MessageBusConfig struct {
	// PublishHost contains the connection infomation for a publishing on zmq
	PublishHost HostInfo
	// SubscribeHost contains the connection infomation for a subscribing on zmq
	SubscribeHost HostInfo
	// Type indicates the message queue platform being used. eg. "zero" for 0mq
	Type string
	// Optional contains all other properties of message bus that is specific to
	// certain concret implementaiton like MQTT's QoS, for example
	Optional map[string]string
}

// HostInfo ...
type HostInfo struct {
	// Host is the hostname or IP address of the messaging broker, if applicable.
	Host string
	// Port defines the port on which to access the message queue.
	Port int
	// Protocol indicates the protocol to use when accessing the message queue.
	Protocol string
}
