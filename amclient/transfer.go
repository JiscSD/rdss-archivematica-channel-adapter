package amclient

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

const transferBasePath = "api/transfer"

type TransferService interface {
	Start(context.Context, *TransferStartRequest) (*Response, error)
	Approve(context.Context, *TransferApproveRequest) (*Response, error)
}

// TransferServiceOp handles communication with the Tranfer related methods of
// the Archivematica API.
type TransferServiceOp struct {
	client *Client
}

var _ TransferService = &TransferServiceOp{}

// Start starts a new transfer.
func (s *TransferServiceOp) Start(ctx context.Context, r *TransferStartRequest) (*Response, error) {
	path := fmt.Sprintf("%s/start_transfer/", transferBasePath)

	req, err := s.client.NewRequest(ctx, "POST", path, r)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil) // TODO: don't ignore response

	return resp, err
}

// Approve approves an existing transfer awaiting for approval.
func (s *TransferServiceOp) Approve(ctx context.Context, r *TransferApproveRequest) (*Response, error) {
	path := fmt.Sprintf("%s/approve/", transferBasePath)

	r.Directory = filepath.Base(r.Directory)

	// TODO: instead of wait we should probably hit the unnaproved/ endpoint and
	// retry until the transfer becomes available as done in automation tools:
	// https://git.io/vHqAu.
	time.Sleep(time.Second * 3)

	req, err := s.client.NewRequest(ctx, "POST", path, r)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil) // TODO: don't ignore response

	return resp, err
}

// TransferStartRequest represents a request to start a transfer.
type TransferStartRequest struct {
	Name  string   `schema:"name"`
	Type  string   `schema:"type"`
	Paths []string `schema:"paths"`
}

// TransferApproveRequest represents a request to approve a transfer.
type TransferApproveRequest struct {
	Type      string `schema:"type"`
	Directory string `schema:"directory"`
}
