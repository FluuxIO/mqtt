package mqtt

import (
	"errors"
	"net"
	"strings"
	"time"
)

// Client is the main structure use to connect as a client on an MQTT
// server.
type Client struct {
	// Store user defined options
	options ClientOptions
	// TCP level connection / can be replace by a TLS session after starttls
	conn net.Conn
}

type Status struct {
	Err error
}

// NewClient generates a new XMPP client, based on Options passed as parameters.
// If host is not specified, the  DNS SRV should be used to find the host from the domainpart of the JID.
// Default the port to 5222.
// TODO: better options checks
func NewClient(options ClientOptions) (c *Client, err error) {
	// TODO: If option address is nil, use the Jid domain to compose the address
	if options.Address, err = checkAddress(options.Address); err != nil {
		return
	}

	c = new(Client)
	c.options = options

	return
}

func checkAddress(addr string) (string, error) {
	var err error
	hostport := strings.Split(addr, ":")
	if len(hostport) > 2 {
		err = errors.New("too many colons in xmpp server address")
		return addr, err
	}

	// Address is composed of two parts, we are good
	if len(hostport) == 2 && hostport[1] != "" {
		return addr, err
	}

	// Port was not passed, we append default MQTT port:
	return strings.Join([]string{hostport[0], "1883"}, ":"), err
}

// Connect initiates asynchronous connection to MQTT server
func (c *Client) Connect() <-chan Status {
	out := make(chan Status)

	go func() {
		var err error
		c.conn, err = net.DialTimeout("tcp", c.options.Address, 5*time.Second)
		if err != nil {
			out <- Status{Err: err}
			return
		}
		out <- Status{}
	}()

	// Connection is ok, we now open MQTT session
	/*	if c.conn, c.Session, err = NewSession(c.conn, c.options); err != nil {
			return err
		}
	*/

	return out
}
