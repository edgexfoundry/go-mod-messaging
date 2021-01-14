/********************************************************************************
 *  Copyright 2020 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package streams

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

func TestNewClientConfiguration(t *testing.T) {
	expectedPassword := "MyPassword"

	tests := []struct {
		name    string
		config  types.MessageBusConfig
		want    OptionalClientConfiguration
		wantErr bool
	}{
		{
			name:    "Create Non Auth OptionalClientConfiguration",
			config:  types.MessageBusConfig{},
			want:    OptionalClientConfiguration{},
			wantErr: false,
		},
		{
			name: "Create Auth uppercase OptionalClientConfiguration",
			config: types.MessageBusConfig{
				Optional: map[string]string{
					"Password": expectedPassword,
				},
			},
			want:    OptionalClientConfiguration{Password: expectedPassword},
			wantErr: false,
		},
		{
			name: "Create Auth lowercase OptionalClientConfiguration",
			config: types.MessageBusConfig{
				Optional: map[string]string{
					"password": expectedPassword,
				},
			},
			// Expect the password not not be set since the name/key is lowercase in the map
			want:    OptionalClientConfiguration{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClientConfiguration(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			if !tt.wantErr {
				require.NoError(t, err)
			}

			assert.Equal(t, got, tt.want)
		})
	}
}
