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
	"math/rand"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

const (
	// Constants for configuration properties provided via the MessageBusConfig's Optional field.
	Username          = "Username"
	Password          = "Password"
	ClientId          = "ClientId"
	Topic             = "Topic"
	Qos               = "Qos"
	KeepAlive         = "KeepAlive"
	Retained          = "Retained"
	ConnectionPayload = "ConnectionPayload"
)

// MQTTClientConfig contains all the configurations for the MQTT client.
type MQTTClientConfig struct {
	BrokerURL string
	MQTTClientOptions
}

// ConnectionOptions contains the connection configurations for the MQTT client.
//
// NOTE: The connection properties resides in its own struct in order to avoid the property being loaded in via
//  reflection during the load process.
type ConnectionOptions struct {
	BrokerURL string
}

// MQTTClientOptions contains the client options which are loaded via reflection
type MQTTClientOptions struct {
	Username          string
	Password          string
	ClientId          string
	Topic             string
	Qos               int
	KeepAlive         int
	Retained          bool
	ConnectionPayload string
}

// CreateMQTTClientConfiguration constructs a MQTTClientConfig based on the provided MessageBusConfig.
func CreateMQTTClientConfiguration(messageBusConfig types.MessageBusConfig) (MQTTClientConfig, error) {
	brokerUrl := messageBusConfig.PublishHost.GetHostURL()
	_, err := url.Parse(brokerUrl)
	if err != nil {
		return MQTTClientConfig{}, err
	}

	mqttClientOptions := CreateMQTTClientOptionsWithDefaults()
	err = load(messageBusConfig.Optional, &mqttClientOptions)
	if err != nil {
		return MQTTClientConfig{}, err
	}

	return MQTTClientConfig{
		BrokerURL:         brokerUrl,
		MQTTClientOptions: mqttClientOptions,
	}, nil
}

// load by reflect to check map key and then fetch the value.
// This function ignores properties that have not been provided from the source. Therefore it is recommended to provide
// a destination struct with reasonable defaults.
//
// NOTE: This logic was borrowed from device-mqtt-go and some additional logic was added to accommodate more types.
// https://github.com/edgexfoundry/device-mqtt-go/blob/a0d50c6e03a7f7dcb28f133885c803ffad3ec502/internal/driver/config.go#L74-L101
func load(config map[string]string, des interface{}) error {
	val := reflect.ValueOf(des).Elem()
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		valueField := val.Field(i)

		val, ok := config[typeField.Name]
		if !ok {
			// Ignore the property if the value is not provided
			continue
		}

		switch valueField.Kind() {
		case reflect.Int:
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return err
			}
			valueField.SetInt(int64(intVal))
		case reflect.String:
			valueField.SetString(val)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			valueField.SetBool(boolVal)
		default:
			return fmt.Errorf("none supported value type %v ,%v", valueField.Kind(), typeField.Name)
		}
	}
	return nil
}

func CreateMQTTClientOptionsWithDefaults() MQTTClientOptions {
	randomClientId := strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100000))
	return MQTTClientOptions{
		Username: "",
		Password: "",
		// Client ID is required or else can cause unexpected errors. This was observed with Eclipse's Mosquito MQTT server.
		ClientId:          randomClientId,
		Topic:             "",
		Qos:               0,
		KeepAlive:         0,
		Retained:          false,
		ConnectionPayload: "",
	}
}
