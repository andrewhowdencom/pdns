package server

import (
	"net"
	"fmt"
	"time"
	"bufio"
	"encoding/binary"
)

const defaultListenPort int = 53
const defaultListenHost string = "0.0.0.0"

// Serve starts up the server listening for DNS connections
func Serve() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", defaultListenHost, defaultListenPort))

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
		go handle(connection)
	}
}

// handle reads the connection looking for the complete DNS packet before passing it off to the assembler
//
// In DNS, the message is not terminated with an EOL character but rather a fixed max length (UDP) or a prefix to the
// message that indicates how long the message will be (TCP).
//
// In this TCP implementation the prefix is
//
// See https://tools.ietf.org/html/rfc1035#section-4.2.2
//
// @todo: Figure out better error handling. DNS should have a defined error return format
func handle(connection net.Conn) {	
	prefix := make([]byte, 2)
	reader := bufio.NewReader(connection)

	// The recommendation from the RFC 1035 is a 2m timeout duration. See:
	// - https://tools.ietf.org/html/rfc1035#section-4.2.2
	timeout, _ := time.ParseDuration("2m")
	err := connection.SetReadDeadline(time.Now().Add(timeout))

	if err != nil {
		fmt.Printf("unable to set a read timeout on the connection: %s", err.Error())
		return	
	}

	// Read the prefix for message length
	_, err = reader.Read(prefix)
	if err != nil {
		fmt.Printf("unable to read the length prefix: %s", err.Error())
		return
	}

	// Get the length of the record
	length := int(binary.BigEndian.Uint16(prefix[:]))

	// Read the body
	payload := make([]byte, length)
	_, err = reader.Read(payload)

	if err != nil {
		fmt.Printf("unable to read data from connection: %s", err.Error())
		return
	}

	fmt.Printf("%x\n", payload)

	// Just close the connection
	connection.Close()
}