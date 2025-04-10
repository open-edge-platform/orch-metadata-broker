/*
* SPDX-FileCopyrightText: (C) 2023 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
*/

package grpc

import (
	"context"

	"github.com/atomix/dazl"
	"github.com/open-edge-platform/orch-library/go/pkg/northbound"
	"github.com/open-edge-platform/orch-library/go/pkg/openpolicyagent"
	"github.com/open-edge-platform/orch-metadata-broker/internal/impl"
	pb "github.com/open-edge-platform/orch-metadata-broker/pkg/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var log = dazl.GetPackageLogger()

// NewService returns a new metadata service
func NewService(opaClient openpolicyagent.ClientWithResponsesInterface) northbound.Service {
	return &Service{
		OpaClient: opaClient,
	}
}

// Service is metadata service.
type Service struct {
	OpaClient openpolicyagent.ClientWithResponsesInterface
}

// Register registers the Service with the gRPC server.
func (s Service) Register(r *grpc.Server) {
	pb.RegisterMetadataServiceServer(r, &Server{opaClient: s.OpaClient})
}

type Server struct {
	opaClient openpolicyagent.ClientWithResponsesInterface
}

const ActiveProjectID = "ActiveProjectID"

// GetActiveProjectID extracts ActiveProjectID metadata from the incoming context
func GetActiveProjectID(ctx context.Context) (*string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "incomplete request: unable to fetch request metadata")
	}

	values := md.Get(ActiveProjectID)
	if len(values) == 0 || values[0] == "" {
		return nil, status.Error(codes.InvalidArgument, "incomplete request: missing ActiveProjectID metadata")
	}

	return &values[0], nil
}

// CreateOrUpdateMetadata creates or updates the metadata, returning updated metadata.
func (s *Server) CreateOrUpdateMetadata(ctx context.Context, request *pb.CreateOrUpdateRequest) (*pb.MetadataResponse, error) {
	projectId, err := GetActiveProjectID(ctx)
	log.Infof("create metadata for project %s: %+v", projectId, request)
	if err != nil {
		return nil, err
	}
	if err := s.authCheckAllowed(ctx, "metadatav1.CreateOrUpdateRequest"); err != nil {
		return nil, err
	}

	for _, m := range request.Body.Metadata {
		if _, err := impl.CreateOrUpdate(projectId, m); err != nil {
			return nil, err
		}
	}

	stored, err := impl.GetSystemMetadata(projectId)
	if err != nil {
		return nil, err
	}

	return &pb.MetadataResponse{Metadata: stored}, nil
}

// Delete removes the specified metadata.
func (s *Server) Delete(ctx context.Context, req *pb.Metadata) (*pb.MetadataResponse, error) {
	projectId, err := GetActiveProjectID(ctx)
	log.Debugf("delete metadata for project %s: %+v", projectId, req)
	if err != nil {
		return nil, err
	}
	if err := s.authCheckAllowed(ctx, "metadatav1.DeleteRequest"); err != nil {
		return nil, err
	}
	res, err := impl.Delete(projectId, req)
	if err != nil {
		return nil, err
	}
	return &pb.MetadataResponse{Metadata: res}, nil
}

// GetMetadata retrieves the current set of metadata.
func (s *Server) GetMetadata(ctx context.Context, empty *emptypb.Empty) (*pb.MetadataResponse, error) {
	projectId, err := GetActiveProjectID(ctx)
	log.Debugf("getting metadata for project %s", projectId)
	if err != nil {
		return nil, err
	}
	if err := s.authCheckAllowed(ctx, "metadatav1.GetRequest"); err != nil {
		return nil, err
	}
	stored, err := impl.GetSystemMetadata(projectId)
	if err != nil {
		return nil, err
	}
	return &pb.MetadataResponse{Metadata: stored}, nil
}

func (s *Server) DeleteProject(ctx context.Context, request *pb.DeleteProjectRequest) (*emptypb.Empty, error) {
	log.Debugf("deleting project %s", request)

	if err := s.authCheckAllowed(ctx, "metadatav1.DeleteProjectRequest"); err != nil {
		return nil, err
	}

	projectId := request.GetId()

	err := impl.DeleteProject(&projectId)

	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
