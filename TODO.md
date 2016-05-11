# TODO

## Done

+ Subscription
+ Fix data race on timer management
+ Message dispatch to client using the library (publish message
  received are dispatched)
+ Basic publish support
+ Support reset of timer when sending a packet from client

## TODO

- Support timeout on PingResp to trigger reconnect
- QOS
- TLS
- Authentication
- Certificate based authentication
- More unit tests
- Example of publish / subscribe sharing Go structures with
  encoding/gob (RPC like)
