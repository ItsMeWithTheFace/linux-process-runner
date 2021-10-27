package api

import (
	"context"
	"log"
	"net"
	"testing"

	pb "github.com/ItsMeWithTheFace/linux-process-runner/api/proto"
	"github.com/ItsMeWithTheFace/linux-process-runner/auth"
	"github.com/ItsMeWithTheFace/linux-process-runner/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/test/bufconn"
)

type TlsAuthTestSuite struct {
	suite.Suite
	lis *bufconn.Listener
}

func (suite *TlsAuthTestSuite) SetupTest() {
	suite.lis = bufconn.Listen(1024 * 1024)
}

func (suite *TlsAuthTestSuite) TestSuccessfulMtlsConnection() {
	serverCreds, err := auth.GetServerTlsCredentials("test_certs/ca1/server.pem", "test_certs/ca1/server.key", "test_certs/ca1/ca1.pem")
	assert.NoError(suite.T(), err, "it should load server credentials")

	clientCreds, err := auth.GetClientTlsCredentials("test_certs/ca1/client.pem", "test_certs/ca1/client.key", "test_certs/ca1/ca1.pem")
	assert.NoError(suite.T(), err, "it should load client credentials")

	s, conn, err := suite.setupServerAndClient(serverCreds, clientCreds)
	assert.NoError(suite.T(), err)

	go func() {
		if err := s.Serve(suite.lis); err != nil {
			log.Fatalf("server exited with error: %v", err)
		}
	}()

	defer conn.Close()
	defer s.Stop()

	c := client.Client{
		JobRunnerServiceClient: pb.NewJobRunnerServiceClient(conn),
	}

	err = c.HandleArgs([]string{"start", "ls"})
	assert.NoError(suite.T(), err, "it should start the job with tls creds")
}

func (suite *TlsAuthTestSuite) TestUnsuccessfulMtlsConnection() {
	serverCreds, err := auth.GetServerTlsCredentials("test_certs/ca1/server.pem", "test_certs/ca1/server.key", "test_certs/ca1/ca1.pem")
	assert.NoError(suite.T(), err, "it should load server credentials")

	clientCreds, err := auth.GetClientTlsCredentials("test_certs/ca1/client.pem", "test_certs/ca1/client.key", "test_certs/ca2/ca2.pem")
	assert.NoError(suite.T(), err, "it should load client credentials")

	s, conn, err := suite.setupServerAndClient(serverCreds, clientCreds)
	assert.NoError(suite.T(), err)

	go func() {
		if err := s.Serve(suite.lis); err != nil {
			log.Fatalf("server exited with error: %v", err)
		}
	}()

	defer conn.Close()
	defer s.Stop()

	c := client.Client{
		JobRunnerServiceClient: pb.NewJobRunnerServiceClient(conn),
	}

	err = c.HandleArgs([]string{"start", "ls"})
	assert.Error(suite.T(), err, "it should not start the job with invalid tls creds")
}

func TestTlsAuthTestSuite(t *testing.T) {
	suite.Run(t, new(TlsAuthTestSuite))
}

func (suite *TlsAuthTestSuite) bufDialer(context.Context, string) (net.Conn, error) {
	return suite.lis.Dial()
}

func (suite *TlsAuthTestSuite) setupServerAndClient(serverCreds credentials.TransportCredentials, clientCreds credentials.TransportCredentials) (*grpc.Server, *grpc.ClientConn, error) {
	s := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.UnaryInterceptor(auth.UnaryAuthInterceptor),
		grpc.StreamInterceptor(auth.StreamAuthInterceptor),
	)
	pb.RegisterJobRunnerServiceServer(s, InitializeJobRunnerServer())

	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(suite.bufDialer), grpc.WithTransportCredentials(clientCreds))

	return s, conn, err
}
