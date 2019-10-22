package app

import (
	"fmt"
	"io"

	"github.com/JiscSD/rdss-archivematica-channel-adapter/version"

	"github.com/spf13/cobra"
)

func NewCmdVersion(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doVersion(out)
		},
	}
}

func doVersion(out io.Writer) error {
	_, err := fmt.Fprintln(out, version.VERSION)
	return err
}
