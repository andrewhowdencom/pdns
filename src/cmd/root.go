package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/dedelala/sysexits"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.pkg.andrewhowden.com/pdns/internal/server"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "pdns",
	Short: "pDNS is a DNS proxy that upconverts DNS to DNS over TLS",
	Long: `pDNS is a reverse proxy that converts DNS into DNS over TLS.
	
It is designed for applications that do not support DNS over TLS so 
that they can ensure the integrity and deliverability of their DNS
responses.

See github.com/andrewhowdencom/pdns`,
	Run: func(cmd *cobra.Command, args []string) {
		err := server.Serve()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("unable to start server")
			os.Exit(sysexits.Software)
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/pdns.yaml)")
	rootCmd.PersistentFlags().StringP("log-level", "v", "warn", "log level for the application (default is warn)")

	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
}

func initConfig() {
	// Read in environment specific configuration\
	viper.SetEnvPrefix("PDNS")
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath("/etc")
		viper.SetConfigName("pdns")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(sysexits.DataErr)
	}

	// Set up logging
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	level, err := log.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		log.SetLevel(log.WarnLevel)
		log.WithFields(log.Fields{
			"log-level": viper.GetString("log-level"),
		}).Warn("Unable to parse log level. Defaulting to warn")
	} else {
		log.SetLevel(level)
	}
}

// Execute instantiates and executes hte application
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("unable to bootstrap the application")
		os.Exit(sysexits.Software)
	}
}
