package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/cmd/consumer"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/cmd/publisher"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/version"
)

const defaultConfig = `# RDSS Archivematica Channel Adapter

[logging]
level = "INFO"

[amclient]
# URL of the Archivematica Dashboard
url = ""
user = ""
key = ""

[s3]
force_path_style = false
insecure_skip_verify = false
endpoint = ""
access_key = ""
secret_key = ""
region = ""

[consumer]
archivematica_shared_dir = "/var/archivematica/sharedDirectory"
backend = "builtin"
dynamodb_tls = true
dynamodb_table = "consumer_storage"
dynamodb_endpoint = ""
dynamodb_region = ""

[publisher]
listen = "0.0.0.0:8000"
tls = false
tls_cert_file = ""
tls_key_file = ""

[broker]
backend = "kinesis"
validation = true

[broker.queues]
main = "main"
invalid = "invalid"
error = "error"

[broker.repository]
backend = "builtin"
dynamodb_tls = true
dynamodb_table = "rdss_am_messages"
dynamodb_endpoint = ""
dynamodb_region = ""

[broker.kinesis]
app_name = "rdss_am"
region = ""
tls = true
endpoint = ""
role_arn = ""
tls_dynamodb = true
endpoint_dynamodb = ""
`

var (
	cfgFile string

	logger = log.StandardLogger()
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "rdss-archivematica-channel-adapter",
	Short: "RDSS Archivematica Channel Adapter",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, setupLogger)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rdss-archivematica-channel-adapter.toml)")

	RootCmd.AddCommand(consumer.Command(logger.WithFields(log.Fields{"cmd": "consumer", "version": version.VERSION})))
	RootCmd.AddCommand(publisher.Command(logger.WithFields(log.Fields{"cmd": "publisher", "version": version.VERSION})))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // Enable ability to specify config file via flag.
		viper.SetConfigFile(cfgFile)
	}

	// Name of config file (without extension).
	viper.SetConfigName(".rdss-archivematica-channel-adapter")

	// Adding home directory as first search path.
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("/etc/archivematica")

	// Read in environment variables that match.
	viper.SetEnvPrefix("RDSS_ARCHIVEMATICA_ADAPTER")
	viper.AutomaticEnv()

	// Set our preferred configuration format.
	viper.SetConfigType("toml")

	// Read our default configuration.
	if err := viper.ReadConfig(strings.NewReader(defaultConfig)); err != nil {
		logger.Fatalln("Cannot read configuration file:", err)
	}

	// If a config file is found, read it in.
	if err := viper.MergeInConfig(); err == nil {
		logger.Infoln("Using config file:", viper.ConfigFileUsed())
	}
}

func setupLogger() {
	var (
		input = viper.GetString("logging.level")
		level = log.InfoLevel
	)

	l, err := log.ParseLevel(input)
	if err != nil {
		logger.Errorln("Not a valid logging level:", input)
	} else {
		level = l
	}

	log.SetLevel(level)
}
