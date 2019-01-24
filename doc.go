/*
Package mqtt implements MQTT client protocol. It can be used as a client library to write MQTT clients in Go.

You can use the MQTT client directly at the low-level and handle connection events in your own code (or ignore them).
If you want to have sane default behaviour for handling reconnect, you can directly rely on the connection manager
struct.
*/
package mqtt // import "gosrc.io/mqtt"
