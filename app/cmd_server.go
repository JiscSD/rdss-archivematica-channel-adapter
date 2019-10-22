package app

import (
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"

	"github.com/JiscSD/rdss-archivematica-channel-adapter/adapter"
	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscSD/rdss-archivematica-channel-adapter/s3"
	"github.com/JiscSD/rdss-archivematica-channel-adapter/version"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCmdServer(logger logrus.FieldLogger, config *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start the application server",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.WithField("v", version.VERSION).Info("Starting server...")
			return doServer(logger, config)
		},
	}
}

func doServer(logger logrus.FieldLogger, config *Config) error {
	var registry *adapter.Registry
	var g run.Group
	{
		var (
			a   *adapter.Adapter
			err error
		)
		a, registry, err = server(logger, config)
		if err != nil {
			return err
		}

		g.Add(func() error {
			a.Run()
			return nil
		}, func(error) {
			a.Stop()
		})
	}
	{
		ln, err := net.Listen("tcp", ":6060")
		if err != nil {
			return err
		}
		logger.WithField("addr", ln.Addr().String()).Info("HTTP server listening")

		g.Add(func() error {
			mux := http.NewServeMux()

			// Health check.
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "OK")
			})

			// Prometheus metrics.
			mux.Handle("/metrics", promhttp.Handler())

			// Profiling data.
			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
			mux.Handle("/debug/pprof/block", pprof.Handler("block"))
			mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
			mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
			mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

			return http.Serve(ln, mux)
		}, func(error) {
			ln.Close()
		})
	}
	{
		cancel := make(chan struct{})

		g.Add(func() error {
			err := interrupt(cancel, registry)
			logger.Warn("Shutting down...")
			return err
		}, func(error) {
			close(cancel)
		})
	}

	return g.Run()
}

func server(logger logrus.FieldLogger, config *Config) (*adapter.Adapter, *adapter.Registry, error) {
	incomingMessages := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "rdss_archivematica_channel_adapter",
		Name:      "incoming_messages_total",
		Help:      "The total number of messages received.",
	})
	prometheus.MustRegister(incomingMessages)

	var dynamodbClient *dynamodb.DynamoDB
	{
		sess, err := awsSession(logger, config.AWS.DynamoDBProfile, config.AWS.DynamoDBEndpoint)
		if err != nil {
			return nil, nil, err
		}
		dynamodbClient = dynamodb.New(sess)
	}

	var brClient *broker.Broker
	{
		sess, err := awsSession(logger, config.AWS.SQSProfile, config.AWS.SQSEndpoint)
		if err != nil {
			return nil, nil, err
		}
		sqsClient := sqs.New(sess)

		sess, err = awsSession(logger, config.AWS.SNSProfile, config.AWS.SNSEndpoint)
		if err != nil {
			return nil, nil, err
		}
		snsClient := sns.New(sess)

		brClient, err = broker.New(
			logger,
			sqsClient, config.Adapter.QueueRecvMainAddr,
			snsClient, config.Adapter.QueueSendMainAddr, config.Adapter.QueueSendInvalidAddr, config.Adapter.QueueSendErrorAddr,
			dynamodbClient, config.Adapter.RepositoryTable,
			config.Adapter.ValidationMode,
			incomingMessages)
		if err != nil {
			return nil, nil, err
		}
	}

	var s3Client s3.ObjectStorage
	{
		sess, err := awsSession(logger, config.AWS.S3Profile, config.AWS.S3Endpoint)
		if err != nil {
			return nil, nil, err
		}
		s3Client = s3.New(sess)
	}

	var storage adapter.Storage
	{
		storage = adapter.NewStorageDynamoDB(dynamodbClient, config.Adapter.ProcessingTable)
	}

	var registry *adapter.Registry
	{
		var err error
		logger := logger.WithField("component", "registry")
		registry, err = adapter.NewRegistry(logger, dynamodbClient, config.Adapter.RegistryTable)
		if err != nil {
			return nil, nil, err
		}
	}

	return adapter.New(logger, brClient, s3Client, storage, registry), registry, nil
}

type logrusProxy struct {
	logger logrus.FieldLogger
}

func (l logrusProxy) Log(args ...interface{}) {
	l.logger.WithField("client", "aws").Debug(args...)
}

// awsSession returns a session using NewSessionWithOptions meaning that it
// relies on the SDK defaults but also the user config files and environment.
//
// AWS_S3_FORCE_PATH_STYLE is a made-up environment string that the SDK does
// not look up. This could be done via configuration instead but I don't want
// to add more surface to the config layer that what's really needed in prod.
func awsSession(logger logrus.FieldLogger, profile, endpoint string) (*session.Session, error) {
	options := session.Options{}
	if profile != "" {
		options.Profile = profile
	}
	if endpoint != "" {
		options.Config.WithEndpoint(endpoint)
	}
	if res, ok := os.LookupEnv("AWS_S3_FORCE_PATH_STYLE"); ok {
		enabled, _ := strconv.ParseBool(res)
		options.Config.WithS3ForcePathStyle(enabled)
	}
	if logrus.GetLevel() == logrus.DebugLevel {
		// options.Config.WithLogLevel(aws.LogDebug | aws.LogDebugWithRequestErrors | aws.LogDebugWithRequestRetries)
		options.Config.WithCredentialsChainVerboseErrors(true)
	}
	options.Config.WithLogger(logrusProxy{logger: logger})
	return session.NewSessionWithOptions(options)
}
