# go-mod-messaging
[![Build Status](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/go-mod-messaging/job/main/badge/icon)](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/go-mod-messaging/job/main/) [![Code Coverage](https://codecov.io/gh/edgexfoundry/go-mod-messaging/branch/main/graph/badge.svg?token=jyOHuKlGPu)](https://codecov.io/gh/edgexfoundry/go-mod-messaging) [![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/go-mod-messaging)](https://goreportcard.com/report/github.com/edgexfoundry/go-mod-messaging) [![GitHub Latest Dev Tag)](https://img.shields.io/github/v/tag/edgexfoundry/go-mod-messaging?include_prereleases&sort=semver&label=latest-dev)](https://github.com/edgexfoundry/go-mod-messaging/tags) ![GitHub Latest Stable Tag)](https://img.shields.io/github/v/tag/edgexfoundry/go-mod-messaging?sort=semver&label=latest-stable) [![GitHub License](https://img.shields.io/github/license/edgexfoundry/go-mod-messaging)](https://choosealicense.com/licenses/apache-2.0/) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/edgexfoundry/go-mod-messaging) [![GitHub Pull Requests](https://img.shields.io/github/issues-pr-raw/edgexfoundry/go-mod-messaging)](https://github.com/edgexfoundry/go-mod-messaging/pulls) [![GitHub Contributors](https://img.shields.io/github/contributors/edgexfoundry/go-mod-messaging)](https://github.com/edgexfoundry/go-mod-messaging/contributors) [![GitHub Committers](https://img.shields.io/badge/team-committers-green)](https://github.com/orgs/edgexfoundry/teams/go-mod-messaging-committers/members) [![GitHub Commit Activity](https://img.shields.io/github/commit-activity/m/edgexfoundry/go-mod-messaging)](https://github.com/edgexfoundry/go-mod-messaging/commits)

Messaging client library for use by Go implementation of EdgeX micro services.  This project contains the abstract Message Bus interface and an implementation for Redis Pub/Sub, MQTT and NATS.
These interface functions connect, publish, subscribe and disconnect to/from the Message Bus.  For more information see the [MessageBus documentation](https://docs.edgexfoundry.org/latest/microservices/general/messagebus/).

### What is this repository for? ###

* Create new MessageClient
* Connect to the Message Bus
* Public messages to the Message Bus
* Subscribe to and receives messages from the Messsage Bus
* Disconnect from the Message Bus 

### Installation ###

* Make sure you have modules enabled, i.e. have an initialized  go.mod file
* If your code is in your GOPATH then make sure ```GO111MODULE=on``` is set
* Run ```go get github.com/edgexfoundry/go-mod-messaging/v3```
  * This will add the go-mod-messaging to the go.mod file and download it into the module cache

### How to Use ###

This library is used by Go programs for interacting with the Message Bus (i.e. redis).

The Message Bus connection information as well as which implementation to use is stored in the service's toml configuration as:

```yaml
MessageBus:
  Protocol: redis
  Host: localhost
  Port: 6379
  Type: redis
```

#### Additional Configuration
Individual client abstractions allow additional configuration properties which can be provided via configuration file:

```yaml
MessageBus:
  Protocol: tcp
  Host: localhost
  Port: 1883
  Type: mqtt
  Topic: events
  Optional:
    ClientId: MyClient
    Username: MyUsername
    ...
```
Or programmatically in the Optional field of the MessageBusConfig struct. For example,

```go
types.MessageBusConfig{
				Broker: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					"ClientId":          "MyClientID",
					"Username":          "MyUser",
					"Password":          "MyPassword",
					...
				}}

```

**NOTE**  
For complete details on configuration options see the [MessageBus documentation](https://docs.edgexfoundry.org/latest/microservices/general/messagebus/)

### Usage
The following code snippets demonstrate how a service uses this messaging module to create a connection, send messages, and receive messages.

This code snippet shows how to connect to the abstract message bus.

```go
var messageBus messaging.MessageClient

var err error
messageBus, err = msgFactory.NewMessageClient(types.MessageBusConfig{
    Broker:   types.HostInfo{
    Host:     Configuration.MessageBus.Host,
    Port:     Configuration.MessageBus.Port,
    Protocol: Configuration.MessageBus.Protocol,
  },
  Type: Configuration.MessageBus.Type,})

if err != nil {
  LoggingClient.Error("failed to create messaging client: " + err.Error())
}

err = messsageBus.Connect()

if err != nil {
  LoggingClient.Error("failed to connect to message bus: " + err.Error())
}
```

This code snippet shows how to publish a message to the abstract message bus.

```go
...
payload, err := json.Marshal(evt)
...
msgEnvelope := types.MessageEnvelope{
  CorrelationID: evt.CorrelationId,
  Payload:       payload,
  ContentType:   clients.ContentJson,
}

err = messageBus.Publish(msgEnvelope, Configuration.MessageBus.Topic)
```

This code snippet shows how to subscribe to the abstract message bus.

```go
messageBus, err := factory.NewMessageClient(types.MessageBusConfig{
    Broker:   types.HostInfo{
    Host:     Configuration.MessageBus.Host,
    Port:     Configuration.MessageBus.Port,
    Protocol: Configuration.MessageBus.Protocol,
  },
  Type: Configuration.MessageBus.Type,
})

if err != nil {
  LoggingClient.Error("failed to create messaging client: " + err.Error())
  return
}

if err := messageBus.Connect(); err != nil {
  LoggingClient.Error("failed to connect to message bus: " + err.Error())
  return
}

topics := []types.TopicChannel{
    {
      Topic:    Configuration.MessageBus.Topic,
      Messages: messages,
    },
}

err = messageBus.Subscribe(topics, messageErrors)
if err != nil {
  LoggingClient.Error("failed to subscribe for event messages: " + err.Error())
  return
}

```

This code snippet shows how to receive data on the message channel after you have subscribed to the bus.

```go
...

for {
select {
  case e := <-errors:
  // handle errors
  ...
  
  case msgEnvelope := <-messages:
    LoggingClient.Info(fmt.Sprintf("Event received on message queue. Topic: %s, Correlation-id: %s ", Configuration.MessageBus.Topic, msgEnvelope.CorrelationID))
    if msgEnvelope.ContentType != clients.ContentJson {
      LoggingClient.Error(fmt.Sprintf("Incorrect content type for event message. Received: %s, Expected: %s", msgEnvelope.ContentType, clients.ContentJson))
      continue
    }
    str := string(msgEnvelope.Payload)
    event := parseEvent(str)
    if event == nil {
      continue
    }
}
...

```