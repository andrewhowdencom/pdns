package server

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

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
}

func (server Server) handleUDP(conn net.PacketConn, addr net.Addr, buffer []byte) {
	response, err := server.proxy(buffer)

	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Error("unable to proxy request to server")
		return
	}

	conn.WriteTo(response, addr)
}
