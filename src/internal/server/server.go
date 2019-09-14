package server

import (
	"errors"

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
