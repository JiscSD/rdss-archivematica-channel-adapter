package app

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func NewCmdConfig(out io.Writer, config *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print the current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doConfig(out, config)
		},
	}
}

func doConfig(out io.Writer, config *Config) error {
	fmt.Fprintln(out, "\n################################################################# Configuration")
	_, err := fmt.Fprintf(out, "%s", config)
	return err
}
