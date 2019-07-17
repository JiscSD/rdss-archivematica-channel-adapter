package app

import (
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const defaultConfig = `# RDSS Archivematica Channel Adapter

################################## LOGGING ####################################

[logging]

#
# Logging verbosity level.
# Supported values: "DEBUG", "INFO", "WARN", "ERROR", "FATAL" or "PANIC".
#
level = "INFO"

################################## ADAPTER ####################################

[adapter]

#
# Name of the table used to store processing state (DynamoDB).
#
processing_table = "rdss_archivematica_adapter_processing_state"

#
# Name of the table used to store the local data repository (DynamoDB).
#
repository_table = "rdss_archivematica_adapter_local_data_repository"

#
# Name of the table used to store the Archivematica registry (DynamoDB).
#
registry_table = "rdss_archivematica_adapter_registry"

#
# Message validation supports three modes:
#
#   validation="strict"
#   Invalid messages are rejected.
#   The validation errors are logged in DEBUG mode.
#
#   validation="warnings"
#   Invalid messages are processed.
#   The validation errors are logged in DEBUG mode.
#
#   validation="disabled"
#   Message validation will not be performed.
#
#
validation_mode = "strict"

#
# AWS SQS queue URL, e.g. "https://queue.amazonaws.com/80398EXAMPLE/MyQueue".
#
# The adapter will subscribe to this queue.
#
queue_recv_main_addr = ""

#
# AWS SNS topic ARN, e.g. "arn:aws:sqs:us-east-2:444455556666:queue1".
#
# The adapter will publish to this queue.
#
queue_send_main_addr = ""
queue_send_error_addr = ""
queue_send_invalid_addr = ""

################################## AWS ########################################

[aws]

s3_profile = ""
s3_endpoint = ""

dynamodb_profile = ""
dynamodb_endpoint = ""

sqs_profile = ""
sqs_endpoint = ""

sns_profile = ""
sns_endpoint = ""
`

type Config struct {
	v *viper.Viper

	Logging struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logging"`

	Adapter struct {
		RepositoryTable      string `mapstructure:"repository_table"`
		ProcessingTable      string `mapstructure:"processing_table"`
		RegistryTable        string `mapstructure:"registry_table"`
		ValidationMode       string `mapstructure:"validation"`
		QueueRecvMainAddr    string `mapstructure:"queue_recv_main_addr"`
		QueueSendMainAddr    string `mapstructure:"queue_send_main_addr"`
		QueueSendErrorAddr   string `mapstructure:"queue_send_error_addr"`
		QueueSendInvalidAddr string `mapstructure:"queue_send_invalid_addr"`
	} `mapstructure:"adapter"`

	AWS struct {
		S3Profile        string `mapstructure:"s3_profile"`
		S3Endpoint       string `mapstructure:"s3_endpoint"`
		DynamoDBProfile  string `mapstructure:"dynamodb_profile"`
		DynamoDBEndpoint string `mapstructure:"dynamodb_endpoint"`
		SQSProfile       string `mapstructure:"sqs_profile"`
		SQSEndpoint      string `mapstructure:"sqs_endpoint"`
		SNSProfile       string `mapstructure:"sns_profile"`
		SNSEndpoint      string `mapstructure:"sns_endpoint"`
	} `mapstructure:"aws"`
}

func (c Config) Validate() error {
	return nil // TODO
}

func (c Config) String() string {
	tmpfile, err := ioutil.TempFile("", "config.*.toml")
	if err != nil {
		return err.Error()
	}
	err = c.v.WriteConfigAs(tmpfile.Name())
	if err != nil {
		return err.Error()
	}
	blob, err := ioutil.ReadAll(tmpfile)
	if err != nil {
		return err.Error()
	}
	return string(blob)
}

func loadConfig(c *Config) error {
	v := viper.New()

	v.SetEnvPrefix("RDSS_ARCHIVEMATICA_ADAPTER")
	v.AutomaticEnv()

	v.SetConfigName("rdss-archivematica-channel-adapter")
	v.SetConfigType("toml")
	v.AddConfigPath("$HOME/.config/")
	v.AddConfigPath("/etc/archivematica/")

	if configFile != "" {
		v.SetConfigFile(configFile)
	}

	// Read our default configuration.
	if err := v.ReadConfig(strings.NewReader(defaultConfig)); err != nil {
		panic(err) // Not in the user path.
	}

	// Include configuration file provided by the user.
	if err := v.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	if err := v.Unmarshal(&c); err != nil {
		return errors.Wrap(err, "configuration unmarshaling failed")
	}

	if err := c.Validate(); err != nil {
		return errors.Wrap(err, "config did not pass validation")
	}

	c.v = v

	return nil
}
