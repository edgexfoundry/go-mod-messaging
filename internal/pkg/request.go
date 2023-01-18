//
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

package pkg

import (
	"fmt"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
	"github.com/google/uuid"
)

func DoRequest(
	subscribe func(topics []types.TopicChannel, messageErrors chan error) error,
	unsubscribe func(topics ...string) error,
	publish func(message types.MessageEnvelope, topic string) error,
	requestMessage types.MessageEnvelope,
	serviceName string,
	requestTopic string,
	responseTopicPrefix string,
	requestTimeout time.Duration) (*types.MessageEnvelope, error) {
	if len(strings.TrimSpace(requestMessage.RequestID)) == 0 {
		requestMessage.RequestID = uuid.NewString()
	}

	// Format of response topic is <prefix>/<service-name>/<request-id>
	responseTopic := strings.Join([]string{responseTopicPrefix, serviceName, requestMessage.RequestID}, "/")
	errs := make(chan error, 1)
	messages := make(chan types.MessageEnvelope, 1)
	responseTopicChan := types.TopicChannel{
		Topic:    responseTopic,
		Messages: messages,
	}

	// Must create the subscription first so that it is in place when the request is handled and response published back
	err := subscribe([]types.TopicChannel{responseTopicChan}, errs)
	if err != nil {
		return nil, fmt.Errorf("unable to create response subscription: %v", err)
	}

	defer func() {
		_ = unsubscribe(responseTopicChan.Topic)
		close(errs)
		close(messages)
	}()

	err = publish(requestMessage, requestTopic)
	if err != nil {
		return nil, fmt.Errorf("unable to create publish request to %s: %v", requestTopic, err)
	}

	select {
	case <-time.After(requestTimeout):
		return nil, fmt.Errorf("timed out waiting for response on %s topic", responseTopicChan.Topic)

	case err = <-errs:
		return nil, fmt.Errorf("encountered error waiting for response to %s: %v", requestTopic, err)

	case responseMessage := <-messages:
		return &responseMessage, nil
	}
}
