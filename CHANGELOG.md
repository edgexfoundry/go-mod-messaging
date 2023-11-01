
<a name="Messaging Go Mod Changelog"></a>
## Messaging Module (in Go)
[Github repository](https://github.com/edgexfoundry/go-mod-messaging)

## Change Logs for EdgeX Dependencies

- [go-mod-core-contracts](https://github.com/edgexfoundry/go-mod-core-contracts/blob/main/CHANGELOG.md)

## [v3.1.0] - 2023-11-15

### ✨  Features

- Jetstream ExactlyOnce Configuration ([3c55305…](https://github.com/edgexfoundry/go-mod-messaging/commit/3c55305cfd05bef29fbb93d4e3d90757e4d22020))

### 👷 Build

- Upgrade to go 1.21 and linter 1.54.2 ([7b59a25…](https://github.com/edgexfoundry/go-mod-messaging/commit/7b59a25c1822550825b9974141cd735907edc1f3))

### 🧪 Testing

- Remove JetStream Integration Tests ([1a5f5cf…](https://github.com/edgexfoundry/go-mod-messaging/commit/1a5f5cf7ee1694f85f5961da7ec7b42515e04c2b))

## [v3.0.0] - 2023-05-31

### Features ✨
- Update API version to v3 ([#dc1fa9](https://github.com/edgexfoundry/go-mod-messaging/commit/dc1fa98dd5cff36050f0a22e2fc1163a68747014))
  ```text
  BREAKING CHANGE: apiVersion in MessageEnvelope has changed to v3
  ```
- Remove deprecated ZeroMQ implementation ([#1801bc4](https://github.com/edgexfoundry/go-mod-messaging/commits/1801bc4))
  ```text
  BREAKING CHANGE:  ZeroMQ no longer an option for the EdgeX MessageBus.
  ```
- Add Request and Unsubscribe capability. Refactor Command Client to use the new Request API ([#2af8c61](https://github.com/edgexfoundry/go-mod-messaging/commit/2af8c61d0e656fe444bc90452b65213b17b562fd))

### Bug Fixes 🐛

- Fix failing unit tests ([#7bea0f0](https://github.com/edgexfoundry/go-mod-messaging/commits/7bea0f0))
- Assure redis correctly subscribe to topic before receiving ([#ff3b20a](https://github.com/edgexfoundry/go-mod-messaging/commits/ff3b20a))
- Subscribe redis topic for empty suffix to match MQTT multi-level wildcard ([#82827e6](https://github.com/edgexfoundry/go-mod-messaging/commits/82827e6))

### Code Refactoring ♻

- Replace topics from config with new constants ([#364627](https://github.com/edgexfoundry/go-mod-messaging/commit/3646279d0d8422a850dd5d44cc6aff0ea4631ac0))
  ```text
  BREAKING CHANGE: Topics no longer in config. Client factory method has additional baseTopic parameter 
  ```
- Refactor Redis Pub/sub topic wild cards to match that of MQTT ([#5426f9](https://github.com/edgexfoundry/go-mod-messaging/commit/5426f937f3aee4cda3bcc0daa344856af0d8fc62))
  ```text
  BREAKING CHANGE: implement the topic conversion between Redis Pub/sub wild card `*` and MQTT wild cards `+` for single level and `#` for multiple levels
  ```
- Command Client: Use bool types for command parameters to be more consistent ([#8751af3](https://github.com/edgexfoundry/go-mod-messaging/commit/8751af38578a3e010e883831076946c588ab1e84))
  ```text
  BREAKING CHANGE: ds-pushevent and ds-returnevent to use bool true/false instead of yes/no
  ```
- Reduce MessageBus config to have a single host ([#e8e57f3](https://github.com/edgexfoundry/go-mod-messaging/commit/e8e57f3b0af30f535f07ddfc2349a27bb1632bae))
  ```text
  BREAKING CHANGE: Configuration now only needs single broker host info
  ```
- Update module to v3 ([#189c2d2](https://github.com/edgexfoundry/go-mod-messaging/commit/189c2d28ed056c67dc7662a8657f6f9aa31afc7f))
  ```text
  BREAKING CHANGE: Import paths will need to change to v3
  ```
- Add JSON attributes to MessageEnvelope fields ([#7366411](https://github.com/edgexfoundry/go-mod-messaging/commits/7366411))
- Remove new mock ([#77f8ddb](https://github.com/edgexfoundry/go-mod-messaging/commits/77f8ddb))
- Restore nats.go import and comment ([#c20c641](https://github.com/edgexfoundry/go-mod-messaging/commits/c20c641))

### Documentation 📖

- Update Stale README ([#a497a32](https://github.com/edgexfoundry/go-mod-messaging/commits/a497a32))

### Build 👷

- Update to Go 1.20 and linter v1.51.2 ([#362e614](https://github.com/edgexfoundry/go-mod-messaging/commits/362e614))

## [v2.3.0] - 2022-11-09

### Features ✨

- Implement Messaging-based CommandClient ([#bcd50ac](https://github.com/edgexfoundry/go-mod-messaging/commits/bcd50ac))
- Add new fields and factory functions for MessageEnvelope ([#a4f6801](https://github.com/edgexfoundry/go-mod-messaging/commits/a4f6801))
- Allow user to bring their own CA in the client ([#aa046c2](https://github.com/edgexfoundry/go-mod-messaging/commits/aa046c2))
- Add NATS Implementation ([#9dc6314](https://github.com/edgexfoundry/go-mod-messaging/commits/9dc6314))
- Add no_zmq build flag ([#51b75ef](https://github.com/edgexfoundry/go-mod-messaging/commits/51b75ef))

### Bug Fixes 🐛

- Add retry of publish if error is EOF ([#54dc8da](https://github.com/edgexfoundry/go-mod-messaging/commits/54dc8da))
- Don't send invalid message as normal message from MQTT ([#7364efb](https://github.com/edgexfoundry/go-mod-messaging/commits/7364efb))
- Handle repeated error better to not spam logs ([#8658677](https://github.com/edgexfoundry/go-mod-messaging/commits/8658677))

### Build 👷

- Upgrade to Go 1.18 ([#55f9df1](https://github.com/edgexfoundry/go-mod-messaging/commits/55f9df1))

## [v2.2.0] - 2022-05-11

### Features ✨

- Disable use of ZMQ on native windows via conditional builds ([#474eb1f](https://github.com/edgexfoundry/go-mod-messaging/commits/474eb1f))

  ```
  BREAKING CHANGE:
  This disables use of ZMQ when running windows native
  builds.
  ```

### Test

- enable go's race detector ([#9e77d98](https://github.com/edgexfoundry/go-mod-messaging/commits/9e77d98))
- Add mock for MessageClient interface ([#c9ac0b8](https://github.com/edgexfoundry/go-mod-messaging/commits/c9ac0b8))

### Bug Fixes 🐛

- Disallow subscribing to same exact topic multiple times ([#c1db79e](https://github.com/edgexfoundry/go-mod-messaging/commits/c1db79e))
- pass a copy of topic to redis subscriber ([#fbf749f](https://github.com/edgexfoundry/go-mod-messaging/commits/fbf749f))
- fix race condition in redis client tests ([#7d21ea0](https://github.com/edgexfoundry/go-mod-messaging/commits/7d21ea0))
- redis subscriber with multiple topics ([#6ddff74](https://github.com/edgexfoundry/go-mod-messaging/commits/6ddff74))

### Documentation 📖

- Update README and GoDoc ([#ee5679d](https://github.com/edgexfoundry/go-mod-messaging/commits/ee5679d))

### Build 👷

- Add build flag to build w/o messging capability ([#397650c](https://github.com/edgexfoundry/go-mod-messaging/commits/397650c))
- **security:** Enable gosec and default linter set ([#0ec7e55](https://github.com/edgexfoundry/go-mod-messaging/commits/0ec7e55))

## [v2.1.0] - 2021-11-17

### Features ✨

- Enable use of CleanSession MQTT option ([#ed2129b](https://github.com/edgexfoundry/go-mod-messaging/commits/ed2129b))

### Bug Fixes 🐛

- Use Qos and Retained value from configuration for MQTT ([#c395010](https://github.com/edgexfoundry/go-mod-messaging/commits/c395010))

## [v2.0.0] - 2021-06-30
### Features ✨
- Add ReceivedTopic to MessageEnvelope & remove Checksum ([#192d447](https://github.com/edgexfoundry/go-mod-messaging/commits/192d447))
    ```
    BREAKING CHANGE:
    Checksum property has been removed from the MessageEnvelope
    ```
### Bug Fixes 🐛
- Use Redis Pub/Sub which supports topic scheme with wild cards ([#e3da10d](https://github.com/edgexfoundry/go-mod-messaging/commits/e3da10d))
    ```
    BREAKING CHANGE:
    Redis Pub/Sub is not compatible with the previous Redis Streams implementation.  All clients must be using the new implementation in order to properly send and receive messages.
    ```
- Resolve race condition in ZMQ impl when binding to port ([#8f0eb58](https://github.com/edgexfoundry/go-mod-messaging/commits/8f0eb58))
### Code Refactoring ♻
- Rename type for Redis implementation to `redis` ([#3ab17e9](https://github.com/edgexfoundry/go-mod-messaging/commits/3ab17e9))
    ```
    BREAKING CHANGE:
    Type for Redis implementation changed from `redisstreams` to `redis`
    ```
- Add Done() method to MockToken for latest Paho module ([#2f6b1ab](https://github.com/edgexfoundry/go-mod-messaging/commits/2f6b1ab))

<a name="v0.1.29"></a>
## [v0.1.29] - 2020-12-29
### Bug Fixes 🐛
- Add ability to re-subscribe to topics when reconnected ([#e6a09cc](https://github.com/edgexfoundry/go-mod-messaging/commits/e6a09cc))

<a name="v0.1.27"></a>
## [v0.1.27] - 2020-10-19
### Bug Fixes 🐛
- Change CorrelationID key constant to match recent change in edgex-go ([#47c5376](https://github.com/edgexfoundry/go-mod-messaging/commits/47c5376))

<a name="v0.1.26"></a>
## [v0.1.26] - 2020-10-09
### Bug Fixes 🐛
- **redisstreams:** Properly set Password via reflection when found in Options ([#f0ba16a](https://github.com/edgexfoundry/go-mod-messaging/commits/f0ba16a))

<a name="v0.1.17"></a>
## [v0.1.17] - 2020-04-03
### Bug
- **race:** Fixed races. ([#b61e84a](https://github.com/edgexfoundry/go-mod-messaging/commits/b61e84a))

