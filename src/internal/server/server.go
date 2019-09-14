package server

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

// Server is an entity reference representing the server
type Server struct {
	Configuration *Configuration
}

// Configuration is an entity that modifies the server behaviour
type Configuration struct {
	// Upstream is the resolver that this proxy will connect to
	Upstream *Host

	// Listen is the address on which this server will listen
	Listen *Host
}

// Host configuration for binding or addressing servers
type Host struct {
	// An IP that represents a host to listen to or to address
	IP string

	// Port to listen for tp to address to
	Port uint16
}

// New returns a new server. In the case that configuration variabels are not defined, set sane defaults.
func New(configuration *Configuration) *Server {
	return &Server{
		Configuration: &Configuration{
			Upstream: &Host{
				IP:   "8.8.8.8",
				Port: 853,
			},
			Listen: &Host{
				IP:   "127.0.0.1",
				Port: 53,
			},
		},
	}
}

// Serve starts up the server listening for DNS connections
func (server Server) Serve() error {
	listener, err := net.Listen(
		"tcp",
		fmt.Sprintf("%s:%d", server.Configuration.Listen.IP, server.Configuration.Listen.Port),
	)

	if err != nil {
		return fmt.Errorf("unable to start listener: %s", err.Error())
	}

	defer listener.Close()

	// Start the connection handler event loop
	for {
		// Read incoming connections off
		connection, err := listener.Accept()

		if err != nil {
			return fmt.Errorf("error accepting connection: %s", err.Error())
		}

		// Handle the connection. Each connection is handled in its own goroutine and assumed to deal with the
		// connection within that context. This allows multiple connections to be executed in parallel
		go server.proxy(connection)
	}
}

func (server Server) proxy(inConn net.Conn) {
	// Set up the connection
	// The recommendation from the RFC 1035 is a 2m timeout duration. See:
	// - https://tools.ietf.org/html/rfc1035#section-4.2.2
	timeout, _ := time.ParseDuration("2m")
	err := inConn.SetReadDeadline(time.Now().Add(timeout))

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("unable to set a read timeout on the connection")

		return
	}

	conf := &tls.Config{}
	outConn, err := tls.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", server.Configuration.Upstream.IP, server.Configuration.Upstream.Port),
		conf,
	)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("unable to connect to upstream resolver")
	}

	defer outConn.Close()

	// Query the upstream
	query, err := read(inConn, true)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("unable to unpack incoming message")
	}

	// Testing
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(query)))

	_, err = outConn.Write(append(length, query...))

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to write to upstream tls socket")
	}

	// Calculate response length
	response, err := read(outConn, true)
	binary.BigEndian.PutUint16(length, uint16(len(response)))

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to read back from connection")
	}

	inConn.Write(append(length, response...))
}

// read extracts the query byte array from a connection object, such as an incoming connection
func read(conn net.Conn, hasLengthPrefix bool) ([]byte, error) {
	if hasLengthPrefix == false {
		return []byte{}, errors.New("reading non length prefixed responses is currently not supported")
	}

	prefix := make([]byte, 2)
	reader := bufio.NewReader(conn)

	// Read the prefix for message length
	_, err := reader.Read(prefix)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to read the length prefix: %s", err.Error())
	}

	// Get the length of the record
	length := int(binary.BigEndian.Uint16(prefix))

	// Read the body
	payload := make([]byte, length)
	_, err = reader.Read(payload)

	if err != nil {
		return []byte{}, fmt.Errorf("unable to read data from connection: %s", err.Error())
	}

	return payload, nil
}
