package api

import (
	"context"
	"math/big"
	"testing"

	"github.com/ItsMeWithTheFace/linux-process-runner/api/proto"
	"github.com/ItsMeWithTheFace/linux-process-runner/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type JobRunnerServerTestSuite struct {
	suite.Suite
	server *JobRunnerServer
}

func (suite *JobRunnerServerTestSuite) SetupTest() {
	suite.server = InitializeJobRunnerServer()
}

func (suite *JobRunnerServerTestSuite) TestUnauthorizedJobAction() {
	mockContext := context.WithValue(context.Background(), auth.ClientIDKey, big.NewInt(123))
	otherMockContext := context.WithValue(context.Background(), auth.ClientIDKey, big.NewInt(456))
	output, err := suite.server.StartJob(mockContext, &proto.JobStartRequest{
		Command: "ls",
	})

	_, err = suite.server.StopJob(otherMockContext, &proto.JobStopRequest{Id: output.Id})
	s, ok := status.FromError(err)
	assert.True(suite.T(), ok, "it should be a grpc error")
	assert.Equal(suite.T(), codes.PermissionDenied, s.Code(), "it should be a permission denied error")
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(JobRunnerServerTestSuite))
}
