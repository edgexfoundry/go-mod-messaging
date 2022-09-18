nats

This package aims to provide EdgeX MessageBus bindings for [NATS](https://nats.io/) - both JetStream and core. JetStream
is probably ideal for most scenarios as it offers persistence and QoS levels similar to MQTT. Core NATS may be useful
for some high throughput publish scenarios as publishing clients do not require acknowledgement.

Note that this package is optionally built with the tag `include_nats_messaging`.

## Why NATS?

NATS is attractive as a messaging solution for EdgeX for a variety of reasons.  It is a generally lightweight protocol that could even be suitable for on-device use, especially when publishing via core NATS.  For devices that use MQTT - it provides [MQTT Client Access](https://docs.nats.io/running-a-nats-service/configuration/mqtt) to underlying JetStream subjects.  And it provides protocol level headers that enable conveying the EdgeX message envelope with no additional marshaling (JSON marshaling / unmarshaling degrades pretty poorly with message size).
## Supported Options

This library does not intend to support all NATS options, and some may be beyond the scope of the edgex message bus
anyway. A good example is that most ack/nack options are not really valid because no ack/nack facility is exposed to the
messagebus trigger via `go-mod-messaging`.  A [custom trigger](https://docs.edgexfoundry.org/2.0/microservices/application/Triggers/#custom-triggers) may be an
option if finer grained control of the NATS configuration is needed.

All option are set via `MessageBusConfig.Optional` and keys are case sensitive .

### EdgeX
| Option        | Description                                                          |
|---------------|----------------------------------------------------------------------|
| Format        | format to use for envelope transport.  Options are 'nats' or 'json'. |

### Connect

| Option               | Description                                                                                                                                                                                                                                    |
|----------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| ClientId             | The id to use to identify the client.  Maps to the `Name` option in nats.go                                                                                                                                                                    |
| ConnectTimeout       | Timeout for NATS connection - an integer representing seconds **(default: 1s)**                                                                                                                                                                |
| RetryOnFailedConnect | Retry on connection failure - expects a string representation of a boolean **(default: true)**                                                                                                                                                 |
| Username             | NATS username                                                                                                                                                                                                                                  |
| Password             | Password for authenticating NATS user                                                                                                                                                                                                          |
| CertFile             | TLS certificate file path                                                                                                                                                                                                                      | 
| KeyFile              | TLS private key file path                                                                                                                                                                                                                      | 
| CertPEMBlock         | TLS certificate as string                                                                                                                                                                                                                      | 
| KeyPEMBlock          | TLS private key as string                                                                                                                                                                                                                      | 
| SkipCertVerify       | Skip client-side certificate verification (not recommended for production use).                                                                                                                                                                |
| AutoProvision        | automatically provision NATS streams. (**JetStream only**)                                                                                                                                                                                     |
| Durable              | Specifies a durable consumer should be used with the given name.  Note that if a durable consumer with the specified name does not exist it will be considered ephemeral and deleted by the client on drain / unsubscribe (**JetStream only**) |
| Subject              | Specifies the subject to use for subscriptions and stream autoprovisioning (Required, **JetStream only**, durable will be favored on subscription.)                                                                                            |


### Subscribe
| Option     | Description                                                                                                                                                                                                                                                                       |
|------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Deliver    | Specifies delivery mode for subscriptions - options are "new", "all", "last" or "lastpersubject".  See the [NATS documentation](https://docs.nats.io/nats-concepts/jetstream/consumers#deliverpolicy-optstartseq-optstarttime) for more detail (**JetStream only, default: new**) |
| QueueGroup | Specifies a queue group to distribute messages from a stream to a pool of worker services.                                                                                                                                                                                        |

### Publish
| Option                  | Description                                                                            |
|-------------------------|----------------------------------------------------------------------------------------|
| DefaultPubRetryAttempts | Number of times to attempt to retry on failed publish (**JetStream only, default: 2**) |

