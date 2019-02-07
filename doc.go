/*
Package mqtt implements MQTT client protocol. It can be used as a client library to write MQTT clients in Go.

You can use the MQTT client directly at the low-level and handle connection events in your own code (or ignore them).
If you want to have sane default behaviour for handling reconnect, you can directly rely on the connection manager
struct.

The messages are received on a message channel. The channel can be buffered. The main goal of the channel is to handle
back pressure and make sure the client will not read message faster than it is able to process.
*/
package mqtt // import "gosrc.io/mqtt"
