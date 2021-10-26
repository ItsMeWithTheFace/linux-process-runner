package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/ItsMeWithTheFace/linux-process-runner/api"
	pb "github.com/ItsMeWithTheFace/linux-process-runner/api/proto"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8080))
	if err != nil {
		log.Printf("failed to listen: %v", err)
		os.Exit(1)
	}
	// TODO: add TLS config
	grpcServer := grpc.NewServer()
	pb.RegisterJobRunnerServiceServer(grpcServer, api.InitializeJobRunnerServer())

	log.Println("starting server...")

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Printf("failed to serve: %v", err)
		os.Exit(1)
	}
}
