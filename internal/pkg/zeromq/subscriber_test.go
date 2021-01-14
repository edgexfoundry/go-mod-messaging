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

package zeromq

import (
	"testing"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

func TestSubscriberInit(t *testing.T) {
	zeromqSubscriber := zeromqSubscriber{}

	msqURL := "tcp://localhost:5690"
	aTopic := types.TopicChannel{Topic: "test", Messages: nil}
	if err := zeromqSubscriber.init(msqURL, &aTopic); err != nil {
		t.Fatalf("cannot init subscriber with URL %s and topic %v", msqURL, aTopic)
	}
}

func TestSubscriberInitNullTopic(t *testing.T) {
	zeromqSubscriber := zeromqSubscriber{}

	msqURL := "tcp://localhost:5690"

	if err := zeromqSubscriber.init(msqURL, nil); err != nil {
		t.Fatalf("cannot init subscriber with URL %s and null topic", msqURL)
	}
}
