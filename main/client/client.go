package main

import (
	"flag"
	"log"

	pb "github.com/ItsMeWithTheFace/linux-process-runner/api/proto"
	"github.com/ItsMeWithTheFace/linux-process-runner/auth"
	"github.com/ItsMeWithTheFace/linux-process-runner/handlers"
	"google.golang.org/grpc"
)

func main() {
	cert := flag.String("cert", "certs/client.pem", "path to the client cert's public key")
	certKey := flag.String("cert-key", "certs/client.key", "path to the client cert's private key")
	caCert := flag.String("ca-cert", "certs/ca.pem", "path to the CA's public key")

	flag.Parse()

	tlsCreds, err := auth.GetClientTlsCredentials(*cert, *certKey, *caCert)
	if err != nil {
		log.Fatalf("could not load tls creds: %s", err.Error())
	}

	// TODO: use configurable server address and port
	conn, err := grpc.Dial("0.0.0.0:8080", grpc.WithTransportCredentials(tlsCreds))

	if err != nil {
		log.Fatalf("could not connect to host: %s", err.Error())
	}

	client := &handlers.Client{
		JobRunnerServiceClient: pb.NewJobRunnerServiceClient(conn),
	}

	err = client.HandleArgs(flag.Args())
	if err != nil {
		log.Fatalf("error handling command args: %s, err: %s", flag.Args(), err.Error())
	}
}
