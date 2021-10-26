package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/ItsMeWithTheFace/linux-process-runner/api"
	pb "github.com/ItsMeWithTheFace/linux-process-runner/api/proto"
	"github.com/ItsMeWithTheFace/linux-process-runner/auth"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()
	// TODO: make address and port configurable
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 8080))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	tlsCreds, err := auth.GetServerTlsCredentials("certs/server.pem", "certs/server.key", "certs/ca.pem")
	if err != nil {
		log.Printf("failed to load tls creds: %v", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(tlsCreds),
		grpc.UnaryInterceptor(auth.UnaryAuthInterceptor),
		grpc.StreamInterceptor(auth.StreamAuthInterceptor),
	)
	pb.RegisterJobRunnerServiceServer(grpcServer, api.InitializeJobRunnerServer())

	log.Println("starting server...")

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
