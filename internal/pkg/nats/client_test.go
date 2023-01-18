//
// Copyright (c) 2022 One Track Consulting
// Copyright (c) 2023 Intel Corporation
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
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v3/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/v3/internal/pkg/nats/interfaces"
	mocks2 "github.com/edgexfoundry/go-mod-messaging/v3/internal/pkg/nats/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createTestClient(t *testing.T) *Client {
	sut, err := NewClient(types.MessageBusConfig{
		Broker: types.HostInfo{
			Host:     "localhost",
			Port:     6869,
			Protocol: "tcp",
		},
	})

	require.NoError(t, err)
	return sut
}
func TestClient_Connect(t *testing.T) {
	sut := createTestClient(t)

	t.Run("no connector", func(t *testing.T) {
		err := sut.Connect()

		assert.Error(t, err)
	})

	t.Run("no host info", func(t *testing.T) {
		sut.connect = func(ClientConfig) (interfaces.Connection, error) {
			return nil, fmt.Errorf("err")
		}
		err := sut.Connect()

		assert.Error(t, err)
	})

	t.Run("connection factory errors", func(t *testing.T) {
		sut.connect = func(ClientConfig) (interfaces.Connection, error) {
			return nil, fmt.Errorf("err")
		}

		err := sut.Connect()

		assert.Error(t, err)
	})

	t.Run("happy", func(t *testing.T) {
		sut.connect = func(ClientConfig) (interfaces.Connection, error) {
			return &mocks2.Connection{}, nil
		}

		err := sut.Connect()
		assert.NoError(t, err)
		assert.NotNil(t, sut.connection)

		t.Run("cant connect again", func(t *testing.T) {
			err := sut.Connect()
			assert.Error(t, err)
			assert.Equal(t, "already connected to NATS", err.Error())
		})
	})
}

func TestClient_Publish(t *testing.T) {
	sut := createTestClient(t)

	t.Run("not connected", func(t *testing.T) {
		err := sut.Publish(types.MessageEnvelope{}, "topic")

		assert.Error(t, err)
	})

	t.Run("empty topic returns error", func(t *testing.T) {
		connection := &mocks2.Connection{}
		sut.connection = connection

		err := sut.Publish(types.MessageEnvelope{}, "")

		assert.Error(t, err)
	})

	t.Run("marshal error", func(t *testing.T) {
		connection := &mocks2.Connection{}
		sut.connection = connection

		m := &mocks2.MarshallerUnmarshaller{}
		m.On("Marshal", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("err"))

		sut.m = m

		err := sut.Publish(types.MessageEnvelope{}, "topic")

		assert.Error(t, err)
	})

	t.Run("publish error", func(t *testing.T) {
		connection := &mocks2.Connection{}
		sut.connection = connection

		nm := nats.NewMsg("topic.id")

		connection.On("PublishMsg", nm).Return(fmt.Errorf("err"))

		m := &mocks2.MarshallerUnmarshaller{}
		m.On("Marshal", mock.Anything, mock.Anything).Return(nm, nil)

		sut.m = m

		err := sut.Publish(types.MessageEnvelope{CorrelationID: "id"}, "topic")

		assert.Error(t, err)
	})

	t.Run("happy", func(t *testing.T) {
		connection := &mocks2.Connection{}
		sut.connection = connection

		nm := nats.NewMsg("topic.id")

		connection.On("PublishMsg", nm).Return(nil)

		m := &mocks2.MarshallerUnmarshaller{}
		m.On("Marshal", mock.Anything, mock.Anything).Return(nm, nil)

		sut.m = m

		err := sut.Publish(types.MessageEnvelope{CorrelationID: "id"}, "topic")

		assert.NoError(t, err)
	})
}

func TestClient_Subscribe(t *testing.T) {
	sut := createTestClient(t)

	t.Run("not connected", func(t *testing.T) {
		ce := make(chan error)
		defer close(ce)
		err := sut.Subscribe([]types.TopicChannel{}, ce)

		assert.Error(t, err)
	})

	t.Run("invalid topic", func(t *testing.T) {
		ce := make(chan error)
		defer close(ce)

		err := sut.Subscribe([]types.TopicChannel{{}}, ce)

		assert.Error(t, err)
	})

	t.Run("subscribe err", func(t *testing.T) {
		ce := make(chan error)
		defer close(ce)

		connection := &mocks2.Connection{}
		connection.On("QueueSubscribe", "topic", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("err"))
		err := sut.Subscribe([]types.TopicChannel{{}}, ce)

		assert.Error(t, err)
	})

	t.Run("happy", func(t *testing.T) {
		ce := make(chan error)
		defer close(ce)

		c1 := make(chan types.MessageEnvelope)

		connection := &mocks2.Connection{}

		connection.On("QueueSubscribe", "topic1", "", mock.Anything).Return(nil, nil)
		sut.connection = connection

		m := &mocks2.MarshallerUnmarshaller{}

		goodMsg := nats.NewMsg(uuid.NewString())
		badMsg := nats.NewMsg(uuid.NewString())
		msgErr := fmt.Errorf("err")

		m.On("Unmarshal", goodMsg, mock.Anything).Return(nil)

		m.On("Unmarshal", badMsg, mock.Anything).Return(msgErr)

		sut.m = m

		err := sut.Subscribe([]types.TopicChannel{
			{Topic: "topic1", Messages: c1},
		}, ce)

		assert.NoError(t, err)

		cb, ok := connection.Calls[0].Arguments[2].(nats.MsgHandler)

		require.True(t, ok)

		go func() {
			cb(goodMsg)
		}()

		select {
		case <-c1:
			break
		case <-time.After(5 * time.Second):
			assert.Fail(t, "no message")
		}

		go func() {
			cb(badMsg)
		}()

		select {
		case e := <-ce:
			require.Equal(t, msgErr, e)
		case <-time.After(5 * time.Second):
			assert.Fail(t, "no message error")
		}
	})
}

func TestClient_Unsubscribe(t *testing.T) {
	sut := createTestClient(t)

	connection := &mocks2.Connection{}
	sut.connection = connection

	connection.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

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

	err := sut.Subscribe(topics, errs)
	require.NoError(t, err)

	_, exists := sut.existingSubscriptions["test1"]
	require.True(t, exists)
	_, exists = sut.existingSubscriptions["test2"]
	require.True(t, exists)
	_, exists = sut.existingSubscriptions["test3"]
	require.True(t, exists)

	err = sut.Unsubscribe("test1")
	require.NoError(t, err)

	_, exists = sut.existingSubscriptions["test1"]
	require.False(t, exists)
	_, exists = sut.existingSubscriptions["test2"]
	require.True(t, exists)
	_, exists = sut.existingSubscriptions["test3"]
	require.True(t, exists)

	err = sut.Unsubscribe("test1", "test2", "test3")
	require.NoError(t, err)

	_, exists = sut.existingSubscriptions["test1"]
	require.False(t, exists)
	_, exists = sut.existingSubscriptions["test2"]
	require.False(t, exists)
	_, exists = sut.existingSubscriptions["test3"]
	require.False(t, exists)
}

func TestNewClientWithConnectionFactory(t *testing.T) {
	cfg := types.MessageBusConfig{Broker: types.HostInfo{Host: "xyz", Protocol: "tcp", Port: 50}, Optional: map[string]string{}}
	var connector ConnectNats

	t.Run("no connector", func(t *testing.T) {
		_, err := NewClientWithConnectionFactory(cfg, connector)
		assert.Error(t, err)
	})

	connector = func(ClientConfig) (interfaces.Connection, error) {
		return &mocks2.Connection{}, nil
	}

	t.Run("default", func(t *testing.T) {
		ct, err := NewClientWithConnectionFactory(cfg, connector)
		assert.NoError(t, err)
		assert.IsType(t, &natsMarshaller{}, ct.m)
	})

	t.Run("json", func(t *testing.T) {
		cfg.Optional[pkg.Format] = "json"
		ct, err := NewClientWithConnectionFactory(cfg, connector)
		assert.NoError(t, err)
		assert.IsType(t, &jsonMarshaller{}, ct.m)
	})

	t.Run("nats", func(t *testing.T) {
		cfg.Optional[pkg.Format] = "nats"
		ct, err := NewClientWithConnectionFactory(cfg, connector)
		assert.NoError(t, err)
		assert.IsType(t, &natsMarshaller{}, ct.m)
	})
}

func TestClient_Disconnect(t *testing.T) {
	for _, tt := range []struct {
		name        string
		drainReturn error
		wantErr     assert.ErrorAssertionFunc
	}{
		{"happy", nil, assert.NoError},
		{"drain error", fmt.Errorf("err"), assert.Error},
	} {
		t.Run(tt.name, func(t *testing.T) {
			conn := &mocks2.Connection{}

			conn.On("Drain").Return(tt.drainReturn)

			sut := createTestClient(t)

			sut.connection = conn

			tt.wantErr(t, sut.Disconnect())
		})
	}
}
