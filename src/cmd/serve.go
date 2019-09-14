/*
Copyright © 2019 Andrew Howden <hello@andrewhowden.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"

	"go.pkg.andrewhowden.com/pdns/internal/server"

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
	Run: func(cmd *cobra.Command, args []string) {
		srv := server.New(&server.Configuration{
			Upstream: &server.Host{
				IP: viper.GetString("server.upstream.ip"),
			},
		})
		err := srv.Serve()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("unable to start server")
			os.Exit(sysexits.Software)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Upstream
	serveCmd.PersistentFlags().StringP("upstream-ip", "", "8.8.8.8", "The upstream resolver to send requests to (default: 8.8.8.8)")
	viper.BindPFlag("server.upstream.ip", serveCmd.PersistentFlags().Lookup("upstream-ip"))
}
