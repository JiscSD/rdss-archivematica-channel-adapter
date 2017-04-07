package outbound

import (
	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/outbound/pb"
)

func init() {
	// I believe this would be a great palce to replace the gRPC logger with
	// our own logger. But we need first a logger, one that meets their
	// interface, which should not be a big deal. See for example how this is
	// solved in github.com/docker/swarmkit/log, although they do something
	// I'm not sure I like: passing loggers via contexts - is that bad?
}

type RdssServer struct {
	broker broker.BrokerAPI
	logger *log.Entry
}

func MakeRdssServer(broker broker.BrokerAPI, logger *log.Entry) pb.RdssServer {
	return &RdssServer{broker, logger}
}

func (s *RdssServer) MetadataRead(ctx context.Context, req *pb.MetadataReadRequest) (*pb.MetadataReadResponse, error) {
	if err := s.broker.ReadMetadata(ctx); err != nil {
		s.logger.Error(err)
	}

	return &pb.MetadataReadResponse{}, nil
}

func (s *RdssServer) MetadataCreate(ctx context.Context, req *pb.MetadataCreateRequest) (*pb.MetadataCreateResponse, error) {
	if err := s.broker.CreateMetadata(ctx); err != nil {
		s.logger.Error(err)
	}

	return &pb.MetadataCreateResponse{}, nil
}

func (s *RdssServer) MetadataUpdate(ctx context.Context, req *pb.MetadataUpdateRequest) (*pb.MetadataUpdateResponse, error) {
	if err := s.broker.UpdateMetadata(ctx); err != nil {
		s.logger.Error(err)
	}

	return &pb.MetadataUpdateResponse{}, nil
}

func (s *RdssServer) MetadataDelete(ctx context.Context, req *pb.MetadataDeleteRequest) (*pb.MetadataDeleteResponse, error) {
	if err := s.broker.DeleteMetadata(ctx); err != nil {
		s.logger.Error(err)
	}

	return &pb.MetadataDeleteResponse{}, nil
}
