package server

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

// NewConfiguration returns a bootstrap configuration with defaults
func NewConfiguration() *Configuration {
	return &Configuration{
		Upstream: &Host{
			IP:   "8.8.8.8",
			Port: 853,
		},
		Listen: &Host{
			IP:   "127.0.0.1",
			Port: 53,
		},
	}
}

// Merge takes another configuration object and merges it into the existing configuration for this object
func (c Configuration) Merge(input *Configuration) {

	if input.Upstream != nil && input.Upstream.IP != "" {
		c.Upstream.IP = input.Upstream.IP
	}
}
