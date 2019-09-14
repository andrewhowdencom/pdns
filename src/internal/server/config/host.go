package config

import (
	"errors"
)

// Host configuration for binding or addressing servers
type Host struct {
	// An IP that represents a host to listen to or to address
	IP string

	// Port to listen for tp to address to
	Port uint16

	// Protocol on which to connect to the upstream service
	Protocol string
}

// SetIP allows overriding the host IP
func (h *Host) SetIP(ip string) error {
	h.IP = ip

	return nil
}

// SetPort allows setting the port on which the host will operate
func (h *Host) SetPort(port uint16) error {
	h.Port = port

	return nil
}

// SetProtocol allows setting the host protocol. Validates against the list of "known" protocols.
func (h *Host) SetProtocol(protocol string) error {
	valid := map[string]bool{ProtoUDP: true, ProtoTCP: true, ProtoDoT: true, ProtoHTTPS: true}
	implemented := map[string]bool{ProtoTCP: true, ProtoDoT: true}

	if _, ok := valid[protocol]; !ok {
		return errors.New("protocol is not valid. See server.go for valid protocols")
	}

	if _, ok := implemented[protocol]; !ok {
		return errors.New("protocol has not yet been implemented")
	}

	h.Protocol = protocol

	return nil
}
