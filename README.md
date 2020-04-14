# go-mod-messaging
Messaging client library for use by Go implementation of EdgeX micro services.  This project contains the abstract Message Bus interface and an implementation for ZeroMQ, and MQTT.
These interface functions connect, publish, subscribe and disconnect to/from the Message Bus.

### What is this repository for? ###

* Create new MessageClient
* Commit to the Message Bus
* Public messages to the Message Bus
* Subscribe to and receives messages from the Messsage Bus
* Disconnect from the Message Bus 

### Installation ###

* Make sure you have modules enabled, i.e. have an initialized  go.mod file
* If your code is in your GOPATH then make sure ```GO111MODULE=on``` is set
* Run ```go get github.com/edgexfoundry/go-mod-messaging```
  * This will add the go-mod-messaging to the go.mod file and download it into the module cache

### How to Use ###
This library is used by Go programs for interacting with the Message Bus (i.e. zeroMQ).  If you are using zeroMQ as your message bus be sure to first [install the zeroMQ library](https://github.com/edgexfoundry/edgex-go#zeromq).

The Message Bus connection information as well as which implementation to use is stored in the service's toml configuration as:

Publisher:
```toml
[MessageQueue]
Protocol = 'tcp'
Host = '*'
Port = 5563
Type = 'zero'
Topic = 'events'
```

Subscriber:
```toml
[MessageQueue]
Protocol = 'tcp'
Host = 'localhost'
Port = 5563
Type = 'zero'
Topic = 'events'
```

The MQTT client abstraction allows for the following additional configuration properties:

- Username
- Password
- ClientId
- Topic
- Qos
- KeepAlive
- Retained
- ConnectionPayload
- CertFile
- KeyFile
- CertPEMBlock
- KeyPEMBlock
- SkipCertVerify

Which can be provided via TOML:

```toml
[MessageQueue]
Protocol = 'tcp'
Host = 'localhost'
Port = 1883
Type = 'mqtt'
Topic = 'events'
    [MessageQueue.Optional]
    ClientId = 'MyClient'
    Username = "MyUsername"
    ...
```
Or programmatically in the Optional field of the MessageBusConfig struct. For example,

```go
types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					"ClientId":          "MyClientID",
					"Username":          "MyUser",
					"Password":          "MyPassword",
					...
				}}

```

**NOTE**  
 The best way to construct the `Optional` map is to use the provided [mqttOptionalConfigurationBuilder](./messaging/mqtt/configuration.go) struct which gives the additional benefit of ensuring the expected types for each property is correct.

The following code snippets demonstrate how a service uses this messaging module to create a connection, send messages, and receive messages.

This code snippet shows how to connect to the abstract message bus.

```go
var messageBus messaging.MessageClient

var err error
messageBus, err = msgFactory.NewMessageClient(types.MessageBusConfig{
  PublishHost: types.HostInfo{
  Host:     Configuration.MessageQueue.Host,
  Port:     Configuration.MessageQueue.Port,
  Protocol: Configuration.MessageQueue.Protocol,
  },
  Type: Configuration.MessageQueue.Type,})

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

err = messageBus.Publish(msgEnvelope, Configuration.MessageQueue.Topic)
```

This code snippet shows how to subscribe to the abstract message bus.

```go
messageBus, err := factory.NewMessageClient(types.MessageBusConfig{
  SubscribeHost: types.HostInfo{
  Host:     Configuration.MessageQueue.Host,
  Port:     Configuration.MessageQueue.Port,
  Protocol: Configuration.MessageQueue.Protocol,
  },
  Type: Configuration.MessageQueue.Type,
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
      Topic:    Configuration.MessageQueue.Topic,
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
    LoggingClient.Info(fmt.Sprintf("Event received on message queue. Topic: %s, Correlation-id: %s ", Configuration.MessageQueue.Topic, msgEnvelope.CorrelationID))
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