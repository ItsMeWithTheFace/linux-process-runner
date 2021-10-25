package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/ItsMeWithTheFace/linux-process-runner/api"
	pb "github.com/ItsMeWithTheFace/linux-process-runner/api/proto"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8080))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	// TODO: add TLS config
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterJobRunnerServiceServer(grpcServer, api.InitializeJobRunnerServer())
	grpcServer.Serve(lis)
}
