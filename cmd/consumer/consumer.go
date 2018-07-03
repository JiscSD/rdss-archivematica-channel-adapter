package consumer

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/amclient"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/consumer"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/s3"

	// Backend implementations
	_ "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend/kinesis"

	// Serve runtime profiling data via HTTP
	_ "net/http/pprof"
)

var cmd = &cobra.Command{
	Use:   "consumer",
	Short: "Consumer server (RDSS » Archivematica)",
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

var logger log.FieldLogger

func Command(l log.FieldLogger) *cobra.Command {
	logger = l
	return cmd
}

func start() {
	logger.Infoln("Hello!")
	defer logger.Info("Bye!")

	go func() {
		logger.Errorln(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	br, err := createBrokerClient()
	if err != nil {
		logger.Fatalln(err)
	}

	s3Client, err := createS3Client()
	if err != nil {
		logger.Fatalln(err)
	}

	amSharedDir := viper.GetString("consumer.archivematica_shared_dir")
	amSharedFs := afero.NewBasePathFs(afero.NewOsFs(), amSharedDir)

	quit := make(chan struct{})
	go func() {
		c := consumer.MakeConsumer(
			ctx, logger,
			br, createAmClient(), s3Client, amSharedFs,
			createConsumerStorage(),
		)
		c.Start()

		quit <- struct{}{}
	}()

	// Subscribe to signals and wait
	stopChan := make(chan os.Signal, 2)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan // Block until a signal is received

	logger.Info("Shutting down server...")
	cancel()
	<-quit
}

func createBrokerClient() (*broker.Broker, error) {
	// Our broker has a backend which we need to configure first.
	backendConfig := map[string]string{
		"app_name":          viper.GetString("broker.kinesis.app_name"),
		"region":            viper.GetString("broker.kinesis.region"),
		"tls":               viper.GetString("broker.kinesis.tls"),
		"endpoint":          viper.GetString("broker.kinesis.endpoint"),
		"role_arn":          viper.GetString("broker.kinesis.role_arn"),
		"tls_dynamodb":      viper.GetString("broker.kinesis.tls_dynamodb"),
		"endpoint_dynamodb": viper.GetString("broker.kinesis.endpoint_dynamodb"),
	}
	var opts = []backend.DialOpts{}
	for key, value := range backendConfig {
		opts = append(opts, backend.WithKeyValue(key, value))
	}
	ba, err := backend.Dial("kinesis", opts...)
	if err != nil {
		return nil, err
	}

	// Build broker config.
	repoConfig := &broker.RepositoryConfig{Backend: viper.GetString("broker.repository.backend")}
	brokerConfig := &broker.Config{
		QueueMain:        viper.GetString("broker.queues.main"),
		QueueInvalid:     viper.GetString("broker.queues.invalid"),
		QueueError:       viper.GetString("broker.queues.error"),
		RepositoryConfig: repoConfig,
	}
	if repoConfig.Backend == "dynamodb" {
		repoConfig.DynamoDBEndpoint = viper.GetString("broker.repository.dynamodb_endpoint")
		repoConfig.DynamoDBRegion = viper.GetString("broker.repository.dynamodb_region")
		repoConfig.DynamoDBTable = viper.GetString("broker.repository.dynamodb_table")
		repoConfig.DynamoDBTLS = viper.GetBool("broker.repository.dynamodb_tls")
	}
	brokerConfig.SetValidationMode(viper.GetString("broker.validation"))

	return broker.New(ba, logger, brokerConfig)
}

func createAmClient() *amclient.Client {
	return amclient.NewClient(nil,
		viper.GetString("amclient.url"),
		viper.GetString("amclient.user"),
		viper.GetString("amclient.key"))
}

func createS3Client() (s3.ObjectStorage, error) {
	var (
		ep                 = viper.GetString("s3.endpoint")
		aKey               = viper.GetString("s3.access_key")
		sKey               = viper.GetString("s3.secret_key")
		region             = viper.GetString("s3.region")
		forcePathStyle     = viper.GetBool("s3.force_path_style")
		insecureSkipVerify = viper.GetBool("s3.insecure_skip_verify")
	)

	opts := []s3.ClientOpt{
		s3.SetForcePathStyle(forcePathStyle),
		s3.SetInsecureSkipVerify(insecureSkipVerify),
	}
	if ep != "" {
		opts = append(opts, s3.SetEndpoint(ep))
	}
	if aKey != "" && sKey != "" {
		opts = append(opts, s3.SetKeys(aKey, sKey))
	}
	if region != "" {
		opts = append(opts, s3.SetRegion(region))
	}
	return s3.New(opts...)
}

func createConsumerStorage() consumer.Storage {
	var (
		backend  = viper.GetString("consumer.backend")
		endpoint = viper.GetString("consumer.dynamodb_endpoint")
		region   = viper.GetString("consumer.dynamodb_region")
		table    = viper.GetString("consumer.dynamodb_table")
		tls      = viper.GetBool("consumer.dynamodb_tls")
	)
	if backend == "builtin" {
		return consumer.NewStorageInMemory()
	}
	if backend == "dynamodb" {
		config := aws.NewConfig()
		if region != "" {
			config = config.WithRegion(region)
		}
		if endpoint != "" {
			config = config.WithEndpoint(endpoint)
		}
		config.DisableSSL = aws.Bool(!tls)
		return consumer.NewStorageDynamoDB(
			dynamodb.New(session.Must(session.NewSession(config))),
			table,
		)
	}
	panic("unknown consumer.backend")
}
