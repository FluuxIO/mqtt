# Native Go MQTT Library

[ ![Codeship Status for FluuxIO/mqtt](https://app.codeship.com/projects/75c09d70-d43d-0135-b59a-12b6e6b26eee/status?branch=master)](https://app.codeship.com/projects/262977) [![Build status](https://ci.appveyor.com/api/projects/status/j3ws3b959b5vdg9j?svg=true)](https://ci.appveyor.com/project/mremond/mqtt)
 [![codecov](https://codecov.io/gh/FluuxIO/mqtt/branch/master/graph/badge.svg)](https://codecov.io/gh/FluuxIO/mqtt)

Fluux MQTT is a MQTT v3.1.1 client library written in Go.

The library has been tested with the following MQTT servers:

- Mosquitto

## Short term tasks

Implement support for QOS 1 and 2 (with storage backend interface and default backends).

## Running tests

You can launch unit tests with:

    go test ./mqtt/...

## Setting Mosquitto on OSX for testing

Client library is currently being tested with Mosquitto.

Mosquitto can be installed from homebrew:

```
brew install mosquitto
...
mosquitto has been installed with a default configuration file.
You can make changes to the configuration by editing:
    /usr/local/etc/mosquitto/mosquitto.conf

To have launchd start mosquitto at login:
  ln -sfv /usr/local/opt/mosquitto/*.plist ~/Library/LaunchAgents
Then to load mosquitto now:
  launchctl load ~/Library/LaunchAgents/homebrew.mxcl.mosquitto.plist
Or, if you don't want/need launchctl, you can just run:
  mosquitto -c /usr/local/etc/mosquitto/mosquitto.conf
```

Default config file can be customized in `/usr/local/etc/mosquitto/mosquitto.conf`.
However, default config file should be ok for testing

You can launch Mosquitto broker with command:

```
/usr/local/sbin/mosquitto -c /usr/local/etc/mosquitto/mosquitto.conf
```

The following command can be use to subscribe a client:

```
mosquitto_sub -v -t 'test/topic'
```

You can publish a payload payload on a topic with:

```
mosquitto_pub -t "test/topic" -m "message payload" -q 1
```

## Setting Mosquitto for testing on Windows 10

After you have install official Mosquitto build from main site, you can run the broker with command:

```
.\mosquitto.exe -v -c .\mosquitto.conf
```

You can subscribe with:

```
.\mosquitto_sub.exe -h 127.0.0.1 -v -t 'test/topic'
```

You can test publish with:

```
.\mosquitto_pub.exe -h 127.0.0.1 -t "test/topic" -m "message payload" -q 1
```
