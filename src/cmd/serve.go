package cmd

import (
	"os"

	"go.pkg.andrewhowden.com/pdns/internal/server"
	"go.pkg.andrewhowden.com/pdns/internal/server/config"

	"github.com/dedelala/sysexits"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the DNS proxy",
	Long: `This command starts the DNS proxy.
	
The proxy should be made available via the defined DNS port 
(default 53) and protocols (default tcp).`,
	Run: start,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Upstream
	serveCmd.PersistentFlags().StringP("upstream-ip", "", "8.8.8.8", "The upstream resolver to send requests to (default: 8.8.8.8)")
	viper.BindPFlag("server.upstream.ip", serveCmd.PersistentFlags().Lookup("upstream-ip"))

	// Listener
	serveCmd.PersistentFlags().StringP("listen-protocol", "", "tcp", "The protocol on which to listen to requests (default: tcp)")
	viper.BindPFlag("server.listen.protocol", serveCmd.PersistentFlags().Lookup("listen-protocol"))

	serveCmd.PersistentFlags().StringP("listen-ip", "", "127.0.0.1", "The protocol on which to listen to requests (default: 127.0.0.1)")
	viper.BindPFlag("server.listen.ip", serveCmd.PersistentFlags().Lookup("listen-ip"))
}

func start(cmd *cobra.Command, args []string) {
	srvCfg := config.New()

	// Configure the server
	if err := srvCfg.Upstream.SetIP(viper.GetString("server.upstream.ip")); err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"upstream-ip": viper.GetString("server.upstream.ip"),
		}).Error("unable to start server")
		os.Exit(sysexits.Software)
	}

	if err := srvCfg.Listen.SetIP(viper.GetString("server.listen.ip")); err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"upstream-ip": viper.GetString("server.upstream.ip"),
		}).Error("unable to start server")
		os.Exit(sysexits.Software)
	}

	if err := srvCfg.Listen.SetProtocol(viper.GetString("server.listen.protocol")); err != nil {
		log.WithFields(log.Fields{
			"error":    err.Error(),
			"protocol": viper.GetString("server.listen.protocol"),
		}).Error("unable to start server")
		os.Exit(sysexits.Software)
	}

	srv := server.New(srvCfg)
	err := srv.Serve()

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("unable to start server")
		os.Exit(sysexits.Software)
	}
}
