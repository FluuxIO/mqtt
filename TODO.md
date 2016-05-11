# TODO

## Done

+ Subscription
+ Fix data race on timer management
+ Message dispatch to client using the library (publish message
  received are dispatched)
+ Basic publish support
+ Support reset of timer when sending a packet from client

## TODO

- Implement unsubscribe
- Implement disconnect
- Use context to clean data flow (https://www.youtube.com/watch?v=3EW1hZ8DVyw&list=PL2ntRZ1ySWBf-_z-gHCOR2N156Nw930Hm)
- Support timeout on PingResp to trigger reconnect
- Reconnect with backoff strategy
- QOS
- TLS
- Authentication
- Certificate based authentication
- More unit tests
- Example of publish / subscribe sharing Go structures with
  encoding/gob (RPC like)
