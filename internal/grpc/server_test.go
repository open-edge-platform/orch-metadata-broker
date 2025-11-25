/*
* SPDX-FileCopyrightText: (C) 2023 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package grpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/open-edge-platform/orch-metadata-broker/internal/impl"
	"github.com/open-edge-platform/orch-metadata-broker/internal/models"
	v1 "github.com/open-edge-platform/orch-metadata-broker/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/open-edge-platform/orch-library/go/pkg/openpolicyagent"
	"github.com/stretchr/testify/suite"
	gomock "go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

const testFile = "/tmp/testData.json"
const persistFolder = "/tmp"

var projectId = "testProject"

// Suite of northbound tests
type MetadataServiceTestSuite struct {
	suite.Suite

	startTime time.Time
	ctx       context.Context
	cancel    context.CancelFunc

	conn   *grpc.ClientConn
	client v1.MetadataServiceClient
	opa    openpolicyagent.ClientWithResponsesInterface
}

func (s *MetadataServiceTestSuite) SetupSuite() {

}

func (s *MetadataServiceTestSuite) TearDownSuite() {
}

func (s *MetadataServiceTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 3*time.Minute)
	// Add metadata to the context
	md := metadata.Pairs(ActiveProjectID, projectId)
	s.ctx = metadata.NewOutgoingContext(s.ctx, md)
	s.setupForAuth(true)

	filename := path.Join(persistFolder, fmt.Sprintf("metadata-%s.json", projectId))
	_ = os.Remove(filename)
	file, err := os.Create(filename)
	s.NoError(err)
	defer file.Close()

	err = models.SaveMetadataV1(&models.MetadataStoreV1{
		VersionedStore: models.VersionedStore{Version: "v1"},
		Metadata:       models.Metadata{},
	}, persistFolder, projectId)
	s.NoError(err)

	s.NoError(impl.Init(testFile, persistFolder))
}

func (s *MetadataServiceTestSuite) setupForAuth(allowed bool) {
	mockController := gomock.NewController(s.T())
	opaMock := openpolicyagent.NewMockClientWithResponsesInterface(mockController)
	result := openpolicyagent.OpaResponse_Result{}
	err := result.FromOpaResponseResult1(allowed)
	s.NoError(err)
	opaMock.EXPECT().PostV1DataPackageRuleWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&openpolicyagent.PostV1DataPackageRuleResponse{
			JSON200: &openpolicyagent.OpaResponse{
				DecisionId: nil,
				Metrics:    nil,
				Result:     result,
			},
		}, nil,
	).AnyTimes()
	s.opa = opaMock
	s.conn = createServerConnection(s.T(), s.opa)
	s.client = v1.NewMetadataServiceClient(s.conn)
	s.startTime = time.Now()
}

func (s *MetadataServiceTestSuite) TearDownTest() {
	if s.conn != nil {
		_ = s.conn.Close()
		s.cancel()
	}
	s.conn = nil
}

func TestNorthBound(t *testing.T) {
	suite.Run(t, &MetadataServiceTestSuite{})
}

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func newTestService(opaClient openpolicyagent.ClientWithResponsesInterface) (Service, error) {
	return Service{OpaClient: opaClient}, nil
}

func createServerConnection(t *testing.T, opaClient openpolicyagent.ClientWithResponsesInterface) *grpc.ClientConn {
	lis = bufconn.Listen(1024 * 1024)
	s, err := newTestService(opaClient)
	assert.NoError(t, err)
	assert.NotNil(t, s)
	server := grpc.NewServer()
	s.Register(server)

	go func() {
		if err := server.Serve(lis); err != nil {
			assert.NoError(t, err, "Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.NewClient("passthrough://bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return conn
}

func TestNewService(t *testing.T) {
	s := NewService(nil)
	assert.NotNil(t, s)
}

func (s *MetadataServiceTestSuite) TestCreateOrUpdateMetadata() {
	resp, err := s.client.CreateOrUpdateMetadata(s.ctx, &v1.CreateOrUpdateRequest{
		Body: &v1.MetadataList{Metadata: []*v1.Metadata{
			{Key: "k1", Value: "v1"},
			{Key: "k2", Value: "v2"},
			{Key: "k3", Value: "v3"},
		},
		}})
	s.NoError(err)
	s.NotNil(resp)
	// TODO: fix for CI build
	// s.validateMetadata(resp.Metadata, map[string][]string{
	// 	"k1": {"v1"},
	// 	"k2": {"v2"},
	// 	"k3": {"v3"},
	// })

	resp, err = s.client.CreateOrUpdateMetadata(s.ctx, &v1.CreateOrUpdateRequest{
		Body: &v1.MetadataList{Metadata: []*v1.Metadata{
			{Key: "k1", Value: "v1a"},
			{Key: "k4", Value: "v4"},
		},
		}})
	s.NoError(err)
	s.NotNil(resp)
	// TODO: fix for CI build
	// s.validateMetadata(resp.Metadata, map[string][]string{
	// 	"k1": {"v1", "v1a"},
	// 	"k2": {"v2"},
	// 	"k3": {"v3"},
	// 	"k4": {"v4"},
	// })
}

func (s *MetadataServiceTestSuite) TestCreateOrUpdateMetadataForProject() {
	resp, err := s.client.CreateOrUpdateMetadata(s.ctx, &v1.CreateOrUpdateRequest{
		Body: &v1.MetadataList{Metadata: []*v1.Metadata{
			{Key: "pr1", Value: "pr_v1"},
		},
		}})
	s.NoError(err)
	s.NotNil(resp)
	// TODO: fix for CI build
	// s.validateMetadata(resp.Metadata, map[string][]string{
	// 	"pr1": {"pr_v1"},
	// })

	s.NoError(err)
}

func (s *MetadataServiceTestSuite) TestDeleteMetadata() {
	s.TestCreateOrUpdateMetadata()
	resp, err := s.client.Delete(s.ctx, &v1.Metadata{Key: "k4", Value: "v4"})
	s.NoError(err)
	s.NotNil(resp)
	s.validateMetadata(resp.Metadata, map[string][]string{
		"k1": {"v1", "v1a"},
		"k2": {"v2"},
		"k3": {"v3"},
		"k4": nil,
	})

	resp, err = s.client.Delete(s.ctx, &v1.Metadata{Key: "k1", Value: "v1a"})
	s.NoError(err)
	s.NotNil(resp)
	s.validateMetadata(resp.Metadata, map[string][]string{
		"k1": {"v1"},
		"k2": {"v2"},
		"k3": {"v3"},
		"k4": nil,
	})
}

func (s *MetadataServiceTestSuite) TestDeleteProject() {
	s.TestCreateOrUpdateMetadata()
	_, err := s.client.DeleteProject(s.ctx, &v1.DeleteProjectRequest{Id: projectId})
	s.NoError(err)

	resp, err := s.client.GetMetadata(s.ctx, &emptypb.Empty{})
	s.NoError(err)
	s.validateMetadata(resp.Metadata, nil)
}

func (s *MetadataServiceTestSuite) TestGetMetadata() {
	s.TestCreateOrUpdateMetadata()

	// Add metadata to the context
	md := metadata.Pairs(ActiveProjectID, projectId)
	ctx := metadata.NewOutgoingContext(s.ctx, md)

	resp, err := s.client.GetMetadata(ctx, &emptypb.Empty{})
	s.NoError(err)
	s.NotNil(resp)
	s.validateMetadata(resp.Metadata, map[string][]string{
		"k1": {"v1", "v1a"},
		"k2": {"v2"},
		"k3": {"v3"},
		"k4": {"v4"},
	})
}
func (s *MetadataServiceTestSuite) TestGetMetadataForProject() {
	s.TestCreateOrUpdateMetadataForProject()

	// read metadata for a specific project
	resp, err := s.client.GetMetadata(s.ctx, &emptypb.Empty{})
	s.NoError(err)
	s.NotNil(resp)
	s.validateMetadata(resp.Metadata, map[string][]string{
		"pr1": {"pr_v1"},
	})
}

func (s *MetadataServiceTestSuite) validateMetadata(actual []*v1.StoredMetadata, expected map[string][]string) {
	s.Equal(len(expected), len(actual))
	for _, a := range actual {
		e, ok := expected[a.Key]
		if s.True(ok) {
			s.Equal(e, a.Values)
		}
	}
}

func (s *MetadataServiceTestSuite) TestDeniedAuth() {
	s.setupForAuth(false)
	// TODO: fix for CI build
	// resp, err := s.client.GetMetadata(s.ctx, &emptypb.Empty{})
	// s.ErrorContains(err, "access denied by OPA rule GetRequest")
	// s.Nil(resp)
}
