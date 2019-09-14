package server

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	"go.pkg.andrewhowden.com/pdns/internal/server/config"

	log "github.com/sirupsen/logrus"
)

// Server is an entity reference representing the server
type Server struct {
	Config *config.Server
}

// New returns a new server. In the case that configuration variabels are not defined, set sane defaults.
func New(cfg *config.Server) *Server {
	return &Server{
		Config: cfg,
	}
}

// Serve starts up the server listening for DNS connections
func (server Server) Serve() error {
	log.WithFields(log.Fields{"config": server.Config}).Debug("starting server")

	if server.Config.Listen.Protocol == config.ProtoTCP {
		return server.serveTCP()
	}

	if server.Config.Listen.Protocol == config.ProtoUDP {
		return server.serveUDP()
	}

	return errors.New("cannot start server: protocol unimplemented")
}

func (server Server) serveUDP() error {
	conn, err := net.ListenPacket(
		server.Config.Listen.Protocol,
		fmt.Sprintf("%s:%d", server.Config.Listen.IP, server.Config.Listen.Port),
	)

	if err != nil {
		return fmt.Errorf("unable to start server: %s", err.Error())
	}

	defer conn.Close()

	for {
		// UDP packets are limited by the RFC to 512 bytes
		// See https://tools.ietf.org/html/rfc1035#section-4.2.1
		buffer := make([]byte, 512)
		length, addr, err := conn.ReadFrom(buffer)

		// Skip failed connections
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Warn("failed to read packet")
			continue
		}

		go server.handleUDP(conn, addr, buffer[:length])

	}

	return nil
}

func (server Server) handleUDP(conn net.PacketConn, addr net.Addr, buffer []byte) {
	response, err := server.proxy(buffer)

	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Error("unable to proxy request to server")
		return
	}

	conn.WriteTo(response, addr)
}

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
