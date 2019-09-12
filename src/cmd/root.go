package cmd

import (
	"fmt"
	"os"

	"github.com/dedelala/sysexits"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		fmt.Println("This functionality has not yet been implemented")
		os.Exit(sysexits.Unavailable)
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/pdns.yaml)")
}

func initConfig() {
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
}

// Execute instantiates and executes hte application
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(sysexits.Software)
	}
}
