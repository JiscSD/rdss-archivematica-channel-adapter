package app

import (
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/adapter"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/amclient"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/s3"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/oklog/run"
	"github.com/pkg/errors"
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
			return doServer(logger, config)
		},
	}
}

func doServer(logger logrus.FieldLogger, config *Config) error {
	var g run.Group
	{
		s, err := server(logger, config)
		if err != nil {
			return err
		}

		g.Add(func() error {
			s.Run()
			return nil
		}, func(error) {
			s.Stop()
		})
	}
	{
		ln, err := net.Listen("tcp", ":6060")
		logger.WithField("addr", ln.Addr().String()).Info("HTTP server listening")
		if err != nil {
			return err
		}

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
			err := interrupt(cancel)
			logger.Warn("Shutting down...")
			return err
		}, func(error) {
			close(cancel)
		})
	}

	return g.Run()
}

func server(logger logrus.FieldLogger, config *Config) (*adapter.Consumer, error) {
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
			return nil, err
		}
		dynamodbClient = dynamodb.New(sess)
	}

	var brClient *broker.Broker
	{
		sess, err := awsSession(logger, config.AWS.SQSProfile, config.AWS.SQSEndpoint)
		if err != nil {
			return nil, err
		}
		sqsClient := sqs.New(sess)

		sess, err = awsSession(logger, config.AWS.SNSProfile, config.AWS.SNSEndpoint)
		if err != nil {
			return nil, err
		}
		snsClient := sns.New(sess)

		brClient, err = broker.New(
			logger,
			sqsClient, config.Adapter.QueueRecvMainAddr,
			snsClient, config.Adapter.QueueSendMainAddr, config.Adapter.QueueSendErrorAddr, config.Adapter.QueueSendInvalidAddr,
			broker.NewRepository(dynamodbClient, config.Adapter.RepositoryTable),
			config.Adapter.ValidationMode,
			incomingMessages)
		if err != nil {
			return nil, err
		}
	}

	var amClient *amclient.Client
	{
		if err := amclient.TransferDir(config.AMClient.TransferDir); err != nil {
			return nil, err
		}
		amClient = amclient.NewClient(
			nil, config.AMClient.URL,
			config.AMClient.User, config.AMClient.Key)
	}

	var s3Client s3.ObjectStorage
	{
		sess, err := awsSession(logger, config.AWS.S3Profile, config.AWS.S3Endpoint)
		if err != nil {
			return nil, err
		}
		s3Client = s3.New(sess)
	}

	var storage adapter.Storage
	{
		storage = adapter.NewStorageDynamoDB(dynamodbClient, config.Adapter.ProcessingTable)
	}

	return adapter.New(logger, brClient, amClient, s3Client, storage), nil
}

type logrusProxy struct {
	logger logrus.FieldLogger
}

func (l logrusProxy) Log(args ...interface{}) {
	l.logger.WithField("client", "aws").Debug(args...)
}

func awsSession(logger logrus.FieldLogger, profile, endpoint string) (*session.Session, error) {
	options := session.Options{}
	if profile != "" {
		options.Profile = profile
	}
	if endpoint != "" {
		options.Config.WithEndpoint(endpoint)
	}
	if logrus.GetLevel() == logrus.DebugLevel {
		// options.Config.WithLogLevel(aws.LogDebug | aws.LogDebugWithRequestErrors | aws.LogDebugWithRequestRetries)
		options.Config.WithCredentialsChainVerboseErrors(true)
	}
	options.Config.WithLogger(logrusProxy{logger: logger})
	return session.NewSessionWithOptions(options)
}

func interrupt(cancel <-chan struct{}) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-c:
		return fmt.Errorf("received signal %s", sig)
	case <-cancel:
		return errors.New("canceled")
	}
}
