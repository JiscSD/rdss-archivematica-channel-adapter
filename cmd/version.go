package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Archivematica RDSS Channel Adapter",
	RunE: func(cmd *cobra.Command, args []string) error {
		printVersion()
		return nil
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

func printVersion() {
	fmt.Println(version.VERSION)
}
