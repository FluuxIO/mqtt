# TODO

## Done

+ Subscription
+ Fix data race on timer management
+ Message dispatch to client using the library (publish message
  received are dispatched)
+ Basic publish support
+ Support reset of timer when sending a packet from client
+ Implement unsubscribe
+ Implement disconnect
+ Make connect synchronous. It is easy in Go to wrap the call to make it asynchronous: http://stackoverflow.com/a/6329459/559289
+ Reconnect with backoff strategy (basic)

## TODO

- We need to setup subscription on reconnect
- Use URL scheme to define connection to server: tcp:// tls://
- Central place for errors definitions
- Use context to clean data flow ? (https://www.youtube.com/watch?v=3EW1hZ8DVyw&list=PL2ntRZ1ySWBf-_z-gHCOR2N156Nw930Hm)
- Support timeout on PingResp to trigger reconnect
- Support subscription based on callbacks or on channels
- QOS
- TLS
- Authentication with username, password. They can be place in URL scheme. tcp://username:password@server 
- Certificate based authentication
- More unit tests
- Example of publish / subscribe sharing Go structures with
  encoding/gob (RPC like)
