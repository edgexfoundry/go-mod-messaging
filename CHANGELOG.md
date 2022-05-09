
<a name="Messaging Go Mod Changelog"></a>
## Messaging Module (in Go)
[Github repository](https://github.com/edgexfoundry/go-mod-messaging)

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

