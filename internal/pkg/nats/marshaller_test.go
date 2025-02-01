//
// Copyright (c) 2022 One Track Consulting
// Copyright (c) 2022-2023 IOTech Ltd
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
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v4/internal/pkg/nats/interfaces"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

var marshallerCases = map[string]interfaces.MarshallerUnmarshaller{
	"json": &jsonMarshaller{},
	"nats": &natsMarshaller{},
}

func TestJSONMarshaller(t *testing.T) {
	suts := []*jsonMarshaller{{}, {ClientConfig{ClientOptions: ClientOptions{ClientId: uuid.NewString(), ExactlyOnce: true}}}}

	for _, sut := range suts {
		t.Run(fmt.Sprintf("ExactlyOnce_%t", sut.opts.ExactlyOnce), func(t *testing.T) {
			pubTopic := uuid.NewString()

			orig := sampleMessage(100)
			expected := orig

			expected.ReceivedTopic = pubTopic // this is set from NATS message

			marshaled, err := sut.Marshal(orig, pubTopic)

			assert.NoError(t, err)

			if sut.opts.ExactlyOnce {
				assert.Equal(t, fmt.Sprintf("%s-%s", sut.opts.ClientId, orig.CorrelationID), marshaled.Header.Get(nats.MsgIdHdr))
			}

			unmarshaled := types.MessageEnvelope{
				QueryParams: make(map[string]string),
			}

			err = sut.Unmarshal(marshaled, &unmarshaled)
			assert.NoError(t, err)
			unmarshaled.Payload, err = types.GetMsgPayload[[]byte](unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, expected, unmarshaled)
		})
	}
}

func TestJSONMarshaller_Unmarshal_Invalid_JSON(t *testing.T) {
	sut := &jsonMarshaller{}

	msg := nats.NewMsg("some.id")

	msg.Data = []byte("{bad")

	err := sut.Unmarshal(msg, &types.MessageEnvelope{})

	require.Error(t, err)
}

func TestNATSMarshaller(t *testing.T) {
	suts := []*natsMarshaller{{}, {ClientConfig{ClientOptions: ClientOptions{ClientId: uuid.NewString(), ExactlyOnce: true}}}}

	for _, sut := range suts {
		t.Run(fmt.Sprintf("ExactlyOnce_%t", sut.opts.ExactlyOnce), func(t *testing.T) {

			pubTopic := uuid.NewString()

			validWithNoQueryParams := sampleMessage(100)
			validWithNoQueryParams.ReceivedTopic = pubTopic
			validWithQueryParams := validWithNoQueryParams
			validWithQueryParams.QueryParams = map[string]string{"foo": "bar"}

			tests := []struct {
				name             string
				envelope         types.MessageEnvelope
				emptyQueryParams bool
			}{
				{"valid", validWithQueryParams, false},
				{"valid - no query parameters", validWithNoQueryParams, true},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					marshaled, err := sut.Marshal(tt.envelope, pubTopic)
					assert.NoError(t, err)
					assert.NotEmpty(t, marshaled.Header.Get(correlationIDHeader))
					assert.NotEmpty(t, marshaled.Header.Get(requestIDHeader))
					assert.Equal(t, common.ApiVersion, marshaled.Header.Get(apiVersionHeader))
					assert.Equal(t, "0", marshaled.Header.Get(errorCodeHeader))
					if tt.emptyQueryParams {
						assert.Empty(t, marshaled.Header.Get(queryParamsHeader))
					} else {
						assert.NotEmpty(t, marshaled.Header.Get(queryParamsHeader))
					}
					if sut.opts.ExactlyOnce {
						assert.Equal(t, fmt.Sprintf("%s-%s", sut.opts.ClientId, tt.envelope.CorrelationID), marshaled.Header.Get(nats.MsgIdHdr))
					}
					unmarshaled := types.MessageEnvelope{}
					err = sut.Unmarshal(marshaled, &unmarshaled)
					assert.NoError(t, err)
					assert.Equal(t, tt.envelope, unmarshaled)
				})
			}
		})
	}
}

func BenchmarkMarshallers(b *testing.B) {
	for _, plSize := range []int{100, 1000, 10000, 20000} {
		for name, sut := range marshallerCases {
			b.Run(fmt.Sprintf("%db-%s", plSize, name), func(b *testing.B) {
				msg := sampleMessage(plSize)

				for i := 0; i < b.N; i++ {
					publishTopic := uuid.NewString()
					marshaled, _ := sut.Marshal(msg, publishTopic)
					unmarshaled := types.MessageEnvelope{}
					_ = sut.Unmarshal(marshaled, &unmarshaled)
				}
			})
		}
	}
}

func sampleMessage(plSize int) types.MessageEnvelope {
	pl := make([]byte, plSize)

	_, _ = rand.Read(pl)

	return types.MessageEnvelope{
		ReceivedTopic: uuid.NewString(),
		CorrelationID: uuid.NewString(),
		Versionable:   commonDTO.NewVersionable(),
		RequestID:     uuid.NewString(),
		ErrorCode:     0,
		Payload:       pl,
		QueryParams:   make(map[string]string),
	}
}
