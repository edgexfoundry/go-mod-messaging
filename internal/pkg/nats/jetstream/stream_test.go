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

package jetstream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_subjectToStreamName(t *testing.T) {
	tests := []struct {
		name  string
		topic string
		want  string
	}{
		{"plain", "topic", "topic"},
		{"nested", "topic.subtopic", "topic_subtopic"},
		{"single level wildcard", "topic.*.subtopic", "topic__subtopic"},
		{"multilevel wildcard", "topic.>", "topic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, subjectToStreamName(tt.topic), "subjectToStreamName(%v)", tt.topic)
		})
	}
}
