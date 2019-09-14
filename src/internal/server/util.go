package server

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

func (server Server) proxy(query []byte) ([]byte, error) {
	conf := &tls.Config{}
	resolver, err := tls.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", server.Config.Upstream.IP, server.Config.Upstream.Port),
		conf,
	)

	if err != nil {
		return make([]byte, 0), fmt.Errorf("cannot connect to upstream resolver: %s", err.Error())
	}

	defer resolver.Close()

	// TCP DNS quries require a length prefix to be applied to the query. See:
	// - https://tools.ietf.org/html/rfc1035#section-4.2.2
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(query)))

	_, err = resolver.Write(append(length, query...))

	if err != nil {
		fmt.Printf("failed to write payload to upstream: %s", err.Error())
	}

	response, err := read(resolver, true)

	if err != nil {
		return make([]byte, 0), fmt.Errorf("unabel to parse response: %s", err.Error())
	}

	return response, nil
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
