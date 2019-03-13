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
	"fmt"

	"github.com/edgexfoundry/go-mod-messaging/pkg/messaging"

	zmq "github.com/pebbe/zmq4"
)

type zeromqSubscriber struct {
	connection *zmq.Socket
	topic      messaging.TopicChannel
	context    *zmq.Context
}

func (subscriber *zeromqSubscriber) init(msgQueueURL string, aTopic *messaging.TopicChannel) (err error) {

	if subscriber.connection == nil {
		subscriber.context, err = zmq.NewContext()

		if err != nil {
			return err
		}

		if subscriber.connection, err = subscriber.context.NewSocket(zmq.SUB); err != nil {
			return err
		}
	}
	if aTopic != nil {
		subscriber.topic = *aTopic
	}

	fmt.Printf("Subscribing to message queue: [%s] ...", msgQueueURL)
	fmt.Println()
	return subscriber.connection.Connect(msgQueueURL)
}
