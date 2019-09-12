package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.pkg.andrewhowden.com/pdns/internal/metadata"
	"github.com/dedelala/sysexits"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the application version",
	Long: `Show a reference that can be used to identify this version of the application.
	
Versions are derived of the git hash of the source code at the time of 
commit, or of a semantic version expressed through git`,
	Run: func(cmd *cobra.Command, args []string) {
		isDetail, err := cmd.Flags().GetBool("detail")

		if err != nil {
			fmt.Println("Unable to determine whether to return detailed version")
			os.Exit(sysexits.Software)
		}

		fmt.Println(version(isDetail))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolP("detail", "", false, "Show more detailed build info")
}

func version(detail bool) string {
	if detail {
		return fmt.Sprintf("%s (%s) %s", metadata.Version, metadata.Hash, metadata.BuildDate)
	}

	return metadata.Version
}