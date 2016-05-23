# Go MQTT

This is an MQTT v3.1.1 client library written in Go.

The library is inspired by my Go XMPP library (gox) and tries to use similar consistant API.

The library has been tested with the following MQTT servers:

- Mosquitto

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
