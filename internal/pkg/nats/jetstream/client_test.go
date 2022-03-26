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
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func Test_parseDeliver(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  func() nats.SubOpt
	}{
		{"all", DeliverAll, nats.DeliverAll},
		{"last", DeliverLast, nats.DeliverLast},
		{"lastpersubject", DeliverLastPerSubject, nats.DeliverLastPerSubject},
		{"new", DeliverNew, nats.DeliverNew},
		{"empty", "", nats.DeliverNew},
		{"not empty or valid", uuid.NewString(), nats.DeliverNew},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//cc := &nats.ConsumerConfig{}

			opt := parseDeliver(tt.input)

			wantAddr := reflect.ValueOf(tt.want).Pointer()
			gotAddr := reflect.ValueOf(opt).Pointer()

			assert.Equal(t, wantAddr, gotAddr)
		})
	}
}
