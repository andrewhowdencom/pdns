package server

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

func (server Server) handleTCP(conn net.Conn, query []byte) {
	length := make([]byte, 2)
	response, err := server.proxy(query)

	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Error("unable to proxy request to server")
		return
	}

	// Calculate the length prefix for TCP connections
	binary.BigEndian.PutUint16(length, uint16(len(response)))

	conn.Write(append(length, response...))
}

func (server Server) serveTCP() error {
	listener, err := net.Listen(
		"tcp",
		fmt.Sprintf("%s:%d", server.Config.Listen.IP, server.Config.Listen.Port),
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

		// The recommendation from the RFC 1035 is a 2m timeout duration. See:
		// - https://tools.ietf.org/html/rfc1035#section-4.2.2
		timeout, _ := time.ParseDuration("2m")
		err = connection.SetReadDeadline(time.Now().Add(timeout))

		if err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Warn("unable to set connection deadline")
			continue
		}

		// Query the upstream
		query, err := read(connection, true)
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Error("unable to unpack incoming message")
			continue
		}

		// Handle the connection. Each connection is handled in its own goroutine and assumed to deal with the
		// connection within that context. This allows multiple connections to be executed in parallel
		go server.handleTCP(connection, query)
	}
}
