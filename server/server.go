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
	// TODO: make address and port configurable
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8080))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// TODO: add TLS config
	grpcServer := grpc.NewServer()
	pb.RegisterJobRunnerServiceServer(grpcServer, api.InitializeJobRunnerServer())

	log.Println("starting server...")

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
