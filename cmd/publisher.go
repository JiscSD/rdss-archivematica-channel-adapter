package cmd

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/outbound"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/outbound/pb"

	// Backend implementations
	_ "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/kinesis"
)

var (
	serverVerbose         bool
	serverInterface       string
	serverPort            int
	serverTLSEnabled      bool
	serverTLSCertFile     string
	serverTLSKeyFile      string
	serverKinesisStream   string
	serverKinesisEndpoint string
)

// publisherCmd represents the publisher command
var publisherCmd = &cobra.Command{
	Use:   "publisher",
	Short: "Outbound server (Archivematica » RDSS)",
	Run:   server,
}

func init() {
	RootCmd.AddCommand(publisherCmd)

	publisherCmd.Flags().BoolVarP(&serverVerbose, "verbose", "v", false, "verbose mode")
	publisherCmd.Flags().StringVarP(&serverInterface, "bind", "b", "127.0.0.1", "interface to which the gRPC server will bind")
	publisherCmd.Flags().IntVarP(&serverPort, "port", "p", 8000, "port on which the gRPC server will listen")
	publisherCmd.Flags().BoolVarP(&serverTLSEnabled, "tls", "", false, "TLS enabled")
	publisherCmd.Flags().StringVarP(&serverTLSCertFile, "tls-cert-file", "", "", "TLS cert file")
	publisherCmd.Flags().StringVarP(&serverTLSKeyFile, "tls-key-file", "", "", "TLS key file")
	publisherCmd.Flags().StringVarP(&serverKinesisStream, "kinesis-stream", "", "", "Kinesis stream")
	publisherCmd.Flags().StringVarP(&serverKinesisEndpoint, "kinesis-endpoint", "", "", "Kinesis endpoint, e.g. https://127.0.0.1:4567")
}

var logger = log.WithFields(log.Fields{"cmd": "publisher"})

func server(cmd *cobra.Command, args []string) {
	if serverVerbose {
		log.SetLevel(log.DebugLevel)
	}

	logger.Info("Hello!")
	defer logger.Info("Bye!")

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	addr := net.JoinHostPort(serverInterface, strconv.Itoa(serverPort))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalln(err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(unaryInterceptor()),
	}
	if serverTLSEnabled {
		creds, err := credentials.NewServerTLSFromFile(serverTLSCertFile, serverTLSCertFile)
		if err != nil {
			logger.Fatalln(err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	grpcServer := grpc.NewServer(opts...)
	rdssServer := outbound.MakeRdssServer(makeBroker(), logger.WithFields(log.Fields{"component": "outbound"}))
	pb.RegisterRdssServer(grpcServer, rdssServer)

	go func() {
		logger.Infof("gRCP server listening on %s", addr)
		grpcServer.Serve(lis)
	}()

	// Subscribe to SIGINT signals and wait
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)
	<-stopChan // Wait for SIGINT

	// Graceful shutdown with a timeout
	logger.Info("Shutting down server...")
	const timeout = time.Second * 3
	c := make(chan struct{})
	go func() {
		defer close(c)
		grpcServer.GracefulStop()
	}()
	select {
	case <-c:
		logger.Info("Server gracefully stopped!")
	case <-time.After(timeout):
		logger.Info("Server timedout when we tried to stop it.")
	}
}

func makeBroker() broker.BrokerAPI {
	var opts []broker.DialOpts
	if serverKinesisStream != "" {
		opts = append(opts, broker.WithKeyValue("stream", serverKinesisStream))
	}
	if serverKinesisEndpoint != "" {
		opts = append(opts, broker.WithKeyValue("endpoint", serverKinesisEndpoint))
	}
	b, err := broker.Dial("kinesis", opts...)
	if err != nil {
		log.Fatalln(err)
	}
	l := logger.WithFields(log.Fields{"component": "broker"})

	return broker.New(b, l)
}

func unaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Debug("gRPC request received")
		return handler(ctx, req)
	}
}
