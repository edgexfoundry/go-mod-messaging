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
	"errors"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoRequest(t *testing.T) {

	var waitTime time.Duration

	expected := types.MessageEnvelope{RequestID: uuid.NewString()}

	unsubscribeFunc := func(topics ...string) error {
		return nil
	}

	unsubscribeErrorFunc := func(topics ...string) error {
		return errors.New("unsubscribe error")
	}

	publishFunc := func(message types.MessageEnvelope, topic string) error {
		return nil
	}

	publishErrorFunc := func(message types.MessageEnvelope, topic string) error {
		return errors.New("publish error")
	}

	subscribeFunc := func(topics []types.TopicChannel, messageErrors chan error) error {
		if len(topics) == 0 {
			return errors.New("no Topics")
		}

		topics[0].Messages <- expected

		return nil
	}

	subscribeErrorFunc := func(topics []types.TopicChannel, messageErrors chan error) error {
		return errors.New("subscribe error")
	}

	subscribeTimeoutFunc := func(topics []types.TopicChannel, messageErrors chan error) error {
		return nil
	}

	subscribeMessageErrorFunc := func(topics []types.TopicChannel, messageErrors chan error) error {
		if len(topics) == 0 {
			return errors.New("no Topics")
		}

		go func() {
			time.Sleep(waitTime)
			messageErrors <- errors.New("subscribe Message Error")
		}()

		return nil
	}

	tests := []struct {
		Name        string
		Subscribe   func(topics []types.TopicChannel, messageErrors chan error) error
		Unsubscribe func(topics ...string) error
		Publish     func(message types.MessageEnvelope, topic string) error
		Timeout     time.Duration
		ExpectError bool
	}{
		{
			Name:        "Valid",
			Subscribe:   subscribeFunc,
			Unsubscribe: unsubscribeFunc,
			Publish:     publishFunc,
			Timeout:     time.Second * 10,
			ExpectError: false,
		},
		{
			Name:        "Subscribe Error",
			Subscribe:   subscribeErrorFunc,
			Unsubscribe: unsubscribeFunc,
			Publish:     publishFunc,
			Timeout:     time.Second * 10,
			ExpectError: true,
		},
		{
			Name:        "Publish Error",
			Subscribe:   subscribeFunc,
			Unsubscribe: unsubscribeFunc,
			Publish:     publishErrorFunc,
			Timeout:     time.Second * 10,
			ExpectError: true,
		},
		{
			Name:        "Valid - Unsubscribe Error",
			Subscribe:   subscribeFunc,
			Unsubscribe: unsubscribeErrorFunc,
			Publish:     publishFunc,
			Timeout:     time.Second * 10,
			ExpectError: false,
		},
		{
			Name:        "Timeout Error",
			Subscribe:   subscribeTimeoutFunc,
			Unsubscribe: unsubscribeFunc,
			Publish:     publishFunc,
			Timeout:     time.Microsecond * 10,
			ExpectError: true,
		},
		{
			Name:        "Message Error",
			Subscribe:   subscribeMessageErrorFunc,
			Unsubscribe: unsubscribeFunc,
			Publish:     publishFunc,
			Timeout:     time.Second * 10,
			ExpectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			actual, err := DoRequest(test.Subscribe, test.Unsubscribe, test.Publish, expected, "test-topic", "edgex/response/my-service", time.Second*5)

			if test.ExpectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, expected, *actual)
		})
	}
}
