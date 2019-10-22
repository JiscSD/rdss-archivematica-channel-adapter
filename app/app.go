package app

import (
	"io"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const defaultLogLevel = logrus.WarnLevel

var (
	configFile     string
	verbosityLevel string
)

func Run(out, stderr io.Writer) error {
	c := RootCommand(out, stderr)
	return c.Execute()
}

func RootCommand(out, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "rdss-archivematica-channel-adapter",
		Short:         "RDSS Archivematica Channel Adapter",
		SilenceErrors: true,
	}

	cmd.SetOutput(out)
	cmd.Root().SilenceUsage = true

	config := &Config{}
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(config); err != nil {
			return err
		}

		if verbosityLevel == "" {
			verbosityLevel = config.Logging.Level
		}
		if err := setUpLogger(out, verbosityLevel, config.Logging.Format); err != nil {
			return err
		}

		return nil
	}

	cmd.AddCommand(NewCmdConfig(out, config))
	cmd.AddCommand(NewCmdVersion(out))
	cmd.AddCommand(NewCmdServer(logrus.WithField("cmd", "server"), config))

	cmd.PersistentFlags().StringVarP(&verbosityLevel, "verbosity", "v", "", "Log level (debug, info, warn, error, fatal, panic)")
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Configuration file")

	return cmd
}

func setUpLogger(out io.Writer, level, format string) error {
	logrus.SetOutput(out)

	if level == "" {
		level = defaultLogLevel.String()
	}
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return errors.Wrap(err, "parsing log level")
	}
	logrus.SetLevel(lvl)

	// For the trace level.
	if lvl > logrus.DebugLevel {
		logrus.SetReportCaller(true)
	}

	switch format {
	case "text":
		// Relying on the default logrus formatter.
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		return errors.Errorf("%s is an unknonw logging format", format)
	}

	return nil
}
