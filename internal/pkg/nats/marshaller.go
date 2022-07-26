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
	"encoding/json"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/nats-io/nats.go"
)

type jsonMarshaller struct{}

func (jm *jsonMarshaller) Marshal(v types.MessageEnvelope, publishTopic string) (*nats.Msg, error) {
	var err error

	subject := topicToSubject(publishTopic)

	out := nats.NewMsg(subject)
	out.Data, err = json.Marshal(v)

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (jm *jsonMarshaller) Unmarshal(msg *nats.Msg, target *types.MessageEnvelope) error {
	topic := subjectToTopic(msg.Subject)

	if err := json.Unmarshal(msg.Data, target); err != nil {
		return err
	}
	target.ReceivedTopic = topic
	return nil
}

const contentTypeHeader = "Content-Type"
const correlationIDHeader = "X-Correlation-ID"

type natsMarshaller struct{}

func (nm *natsMarshaller) Marshal(v types.MessageEnvelope, publishTopic string) (*nats.Msg, error) {
	subject := topicToSubject(publishTopic)

	out := nats.NewMsg(subject)
	out.Data = v.Payload
	out.Header.Set(correlationIDHeader, v.CorrelationID)
	out.Header.Set(contentTypeHeader, v.ContentType)

	return out, nil
}

func (nm *natsMarshaller) Unmarshal(msg *nats.Msg, target *types.MessageEnvelope) error {
	topic := subjectToTopic(msg.Subject)

	target.ReceivedTopic = topic

	target.Payload = msg.Data
	target.CorrelationID = msg.Header.Get(correlationIDHeader)
	target.ContentType = msg.Header.Get(contentTypeHeader)

	return nil
}
