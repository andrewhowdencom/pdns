package config

// The various pipes that DNS information can be sent and received from
//
// See https://en.wikipedia.org/wiki/Domain_Name_System
const (
	ProtoUDP   = "udp"
	ProtoTCP   = "tcp"
	ProtoDoT   = "dot"
	ProtoHTTPS = "doh"
)

// Configuration is an entity that modifies the server behaviour
type Server struct {
	// Upstream is the resolver that this proxy will connect to
	Upstream *Host

	// Listen is the address on which this server will listen
	Listen *Host
}

// NewConfiguration returns a bootstrap configuration with defaults
func New() *Server {
	return &Server{
		Upstream: &Host{
			IP:       "8.8.8.8",
			Port:     853,
			Protocol: ProtoDoT,
		},
		Listen: &Host{
			IP:       "127.0.0.1",
			Port:     53,
			Protocol: ProtoTCP,
		},
	}
}
