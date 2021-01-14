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

package mqtt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v2/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/v2/internal/pkg/mqtt"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

func TestBuilderMethods(t *testing.T) {
	tests := []struct {
		name           string
		builder        *mqttOptionalConfigurationBuilder
		expectedValues map[string]string
	}{
		{
			name:           "AutoReconnect",
			builder:        NewMQTTOptionalConfigurationBuilder().AutoReconnect(true),
			expectedValues: map[string]string{AutoReconnect: "true"},
		},
		{
			name:           "CertFile",
			builder:        NewMQTTOptionalConfigurationBuilder().CertFile("/path/to/some/cert"),
			expectedValues: map[string]string{CertFile: "/path/to/some/cert"},
		},
		{
			name:           "CertPEMBlock",
			builder:        NewMQTTOptionalConfigurationBuilder().CertPEMBlock("/path/to/some/cert"),
			expectedValues: map[string]string{CertPEMBlock: "/path/to/some/cert"},
		},
		{
			name:           "ClientID",
			builder:        NewMQTTOptionalConfigurationBuilder().ClientID("ProvidedClientID"),
			expectedValues: map[string]string{ClientId: "ProvidedClientID"},
		},
		{
			name:           "ConnectTimeout",
			builder:        NewMQTTOptionalConfigurationBuilder().ConnectTimeout(99),
			expectedValues: map[string]string{ConnectTimeout: "99"},
		},
		{
			name:           "KeepAlive",
			builder:        NewMQTTOptionalConfigurationBuilder().KeepAlive(99),
			expectedValues: map[string]string{KeepAlive: "99"},
		},
		{
			name:           "KeyPEMBlock",
			builder:        NewMQTTOptionalConfigurationBuilder().KeyPEMBlock("ProvidedKeyPEMBlock"),
			expectedValues: map[string]string{KeyPEMBlock: "ProvidedKeyPEMBlock"},
		},
		{
			name:           "KeyFile",
			builder:        NewMQTTOptionalConfigurationBuilder().KeyFile("ProvidedKeyFile"),
			expectedValues: map[string]string{KeyFile: "ProvidedKeyFile"},
		},
		{
			name:           "Password",
			builder:        NewMQTTOptionalConfigurationBuilder().Password("ProvidedPassword"),
			expectedValues: map[string]string{Password: "ProvidedPassword"},
		},
		{
			name:           "Qos",
			builder:        NewMQTTOptionalConfigurationBuilder().Qos(1),
			expectedValues: map[string]string{Qos: "1"},
		},
		{
			name:           "Retained",
			builder:        NewMQTTOptionalConfigurationBuilder().Retained(true),
			expectedValues: map[string]string{Retained: "true"},
		},
		{
			name:           "SkipCertVerify",
			builder:        NewMQTTOptionalConfigurationBuilder().SkipCertVerify(true),
			expectedValues: map[string]string{SkipCertVerify: "true"},
		},
		{
			name:           "Username",
			builder:        NewMQTTOptionalConfigurationBuilder().Username("ProvidedUsername"),
			expectedValues: map[string]string{Username: "ProvidedUsername"},
		},
		{
			name: "Multiple Properties of Different Types",
			builder: NewMQTTOptionalConfigurationBuilder().
				Username("ProvidedUsername").
				SkipCertVerify(true).
				KeepAlive(99),
			expectedValues: map[string]string{Username: "ProvidedUsername", SkipCertVerify: "true", KeepAlive: "99"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			observedConfig := test.builder.Build()

			for key, value := range test.expectedValues {
				assert.Equal(t, value, observedConfig[key])
			}

		})
	}
}

func TestClientOptionsIntegration(t *testing.T) {
	protocol := "tcp"
	host := "test.com"
	port := 8080

	expectedConfig := mqtt.MQTTClientConfig{
		BrokerURL: fmt.Sprintf("%s://%s:%d", protocol, host, port),
		MQTTClientOptions: mqtt.MQTTClientOptions{
			Username:       "ProvidedUsername",
			Password:       "ProvidedPassword",
			ClientId:       "ProvidedClientId",
			Qos:            99,
			KeepAlive:      98,
			Retained:       true,
			AutoReconnect:  true,
			ConnectTimeout: 97,
			TlsConfigurationOptions: pkg.TlsConfigurationOptions{
				SkipCertVerify: true,
				CertFile:       "ProvidedCertFile",
				KeyFile:        "ProvidedKeyFile",
				KeyPEMBlock:    "ProvidedKeyPEMBlock",
				CertPEMBlock:   "ProvidedCertPEMBlock",
			},
		},
	}
	optionalConfigurations := NewMQTTOptionalConfigurationBuilder().
		Username(expectedConfig.Username).
		Password(expectedConfig.Password).
		ClientID(expectedConfig.ClientId).
		Qos(expectedConfig.Qos).
		KeepAlive(expectedConfig.KeepAlive).
		Retained(expectedConfig.Retained).
		AutoReconnect(expectedConfig.AutoReconnect).
		ConnectTimeout(expectedConfig.ConnectTimeout).
		SkipCertVerify(expectedConfig.SkipCertVerify).
		CertFile(expectedConfig.CertFile).
		KeyFile(expectedConfig.KeyFile).
		KeyPEMBlock(expectedConfig.KeyPEMBlock).
		CertPEMBlock(expectedConfig.CertPEMBlock).
		Build()

	messageBusConfig := types.MessageBusConfig{
		PublishHost: types.HostInfo{
			Host:     host,
			Port:     port,
			Protocol: protocol,
		},
		Optional: optionalConfigurations,
	}

	observedResult, err := mqtt.CreateMQTTClientConfiguration(messageBusConfig)
	require.NoError(t, err)
	assert.Equal(t, expectedConfig, observedResult)
}
