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
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/edgexfoundry/go-mod-messaging/v2/internal/pkg/nats/interfaces"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var marshallerCases = map[string]interfaces.MarshallerUnmarshaller{
	"json": &jsonMarshaller{},
	"nats": &natsMarshaller{},
}

func TestMarshallers(t *testing.T) {
	for name, sut := range marshallerCases {
		t.Run(name, func(t *testing.T) {
			pubTopic := uuid.NewString()

			orig := sampleMessage(100)
			expected := orig

			expected.ReceivedTopic = pubTopic // this is set from NATS message

			marshaled, err := sut.Marshal(orig, pubTopic)

			assert.NoError(t, err)

			unmarshaled := types.MessageEnvelope{}

			err = sut.Unmarshal(marshaled, &unmarshaled)

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
		Payload:       pl,
		ContentType:   uuid.NewString(),
	}
}
