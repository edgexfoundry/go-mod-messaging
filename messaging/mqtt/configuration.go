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

// Package mqtt provides additional functionality to aid in configuring a MQTT client.
package mqtt

import (
	"strconv"

	"github.com/edgexfoundry/go-mod-messaging/v4/internal/pkg"
)

// mqttOptionalConfigurationBuilder encapsulates the optional configuration data following the builder pattern. Updating
// values are done via the exported builder methods and the final map can be obtained by calling the Build method.
//
// This is the recommended way to create optional parameters for an MQTT client since the underlying structure is a map
// of string : string which can introduce issues with type casting. The builder methods for this struct are created to
// accept the expected types to reduce the chance of incorrect type casting.
type mqttOptionalConfigurationBuilder struct {
	options map[string]string
}

// NewMQTTOptionalConfigurationBuilder constructs a new mqttOptionalConfigurationBuilder.
func NewMQTTOptionalConfigurationBuilder() *mqttOptionalConfigurationBuilder {
	return &mqttOptionalConfigurationBuilder{
		options: make(map[string]string),
	}
}

// Build constructs the optional configuration map based which can be used when creating an MQTT client.
func (mocb *mqttOptionalConfigurationBuilder) Build() map[string]string {
	return mocb.options
}

// AutoReconnect sets the autoReconnect configuration property and returns the builder struct for further updates.
func (mocb *mqttOptionalConfigurationBuilder) AutoReconnect(autoReconnect bool) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.AutoReconnect] = strconv.FormatBool(autoReconnect)
	return mocb
}

// CleanSession sets the CleanSession configuration property and returns the builder struct for further updates.
func (mocb *mqttOptionalConfigurationBuilder) CleanSession(cleanSession bool) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.CleanSession] = strconv.FormatBool(cleanSession)
	return mocb
}

// CertFile sets the certFile configuration property and returns the builder struct for further updates.
func (mocb *mqttOptionalConfigurationBuilder) CertFile(certFile string) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.CertFile] = certFile
	return mocb
}

// CertPEMBlock sets the certPEMBlock configuration property and returns the builder struct for further updates.
func (mocb *mqttOptionalConfigurationBuilder) CertPEMBlock(certPEMBlock string) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.CertPEMBlock] = certPEMBlock
	return mocb
}

// ClientID sets the clientID configuration property and returns the builder struct for further updates.
func (mocb *mqttOptionalConfigurationBuilder) ClientID(clientID string) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.ClientId] = clientID
	return mocb
}

// ConnectTimeout sets the connectionTimeout configuration property in seconds and returns the builder struct for
// further updates.
func (mocb *mqttOptionalConfigurationBuilder) ConnectTimeout(connectionTimeout int) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.ConnectTimeout] = strconv.Itoa(connectionTimeout)
	return mocb
}

// KeepAlive sets the keepAlive duration in seconds configuration property and returns the builder struct for further
// updates.
func (mocb *mqttOptionalConfigurationBuilder) KeepAlive(keepAlive int) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.KeepAlive] = strconv.Itoa(keepAlive)
	return mocb
}

// KeyPEMBlock sets the keyPEMBlock configuration property and returns the builder struct for further updates.
func (mocb *mqttOptionalConfigurationBuilder) KeyPEMBlock(keyPEMBlock string) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.KeyPEMBlock] = keyPEMBlock
	return mocb
}

// KeyFile sets the fileLocation configuration property and returns the builder struct for further updates.
func (mocb *mqttOptionalConfigurationBuilder) KeyFile(fileLocation string) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.KeyFile] = fileLocation
	return mocb
}

// Password sets the password configuration property and returns the builder struct for further updates.
func (mocb *mqttOptionalConfigurationBuilder) Password(password string) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.Password] = password
	return mocb
}

// Qos sets the qos configuration property and returns the builder struct for further updates.
func (mocb *mqttOptionalConfigurationBuilder) Qos(qos int) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.Qos] = strconv.Itoa(qos)
	return mocb
}

// Retained sets the retained configuration property and returns the builder struct for further updates.
func (mocb *mqttOptionalConfigurationBuilder) Retained(retained bool) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.Retained] = strconv.FormatBool(retained)
	return mocb
}

// SkipCertVerify sets the skipCertVerify configuration property and returns the builder struct for further updates.
func (mocb *mqttOptionalConfigurationBuilder) SkipCertVerify(skipCertVerify bool) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.SkipCertVerify] = strconv.FormatBool(skipCertVerify)
	return mocb
}

// Username sets the username configuration property and returns the builder struct for further updates.
func (mocb *mqttOptionalConfigurationBuilder) Username(username string) *mqttOptionalConfigurationBuilder {
	mocb.options[pkg.Username] = username
	return mocb
}
