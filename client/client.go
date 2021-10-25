package main

import (
	"context"
	"flag"
	"fmt"
	"io"

	pb "github.com/ItsMeWithTheFace/linux-process-runner/api/proto"
	"google.golang.org/grpc"
)

type Client struct {
	pb.JobRunnerServiceClient
}

func (c *Client) handleArgs(args []string) {
	if len(args) < 1 {
		fmt.Println("Please provide one of the following commands: [start, stop, get, stream]")
		return
	}
	switch command := args[0]; command {
	case "start":
		c.handleStartJobCommand(context.Background(), args[1], args[2:])
	case "stop":
		c.handleStopJobCommand(context.Background(), args[1])
	case "get":
		c.handleGetJobCommand(context.Background(), args[1])
	case "stream":
		c.handleStreamJobOutputCommand(context.Background(), args[1])
	default:
		fmt.Println("Please provide one of the following commands: [start, stop, get, stream]")
	}
}

func (c *Client) handleStartJobCommand(ctx context.Context, command string, args []string) {
	fmt.Println("starting job")
	_, err := c.JobRunnerServiceClient.StartJob(ctx, &pb.JobStartRequest{Command: command, Arguments: args})

	if err != nil {
		fmt.Print(err.Error())
	}
}

func (c *Client) handleStopJobCommand(ctx context.Context, id string) {
	c.JobRunnerServiceClient.StopJob(ctx, &pb.JobStopRequest{Id: id})
}

func (c *Client) handleGetJobCommand(ctx context.Context, id string) {
	job, err := c.JobRunnerServiceClient.GetJobInfo(ctx, &pb.JobQueryRequest{Id: id})

	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Println(job)
}

func (c *Client) handleStreamJobOutputCommand(ctx context.Context, id string) {
	srv, err := c.JobRunnerServiceClient.StreamJobOutput(ctx, &pb.JobQueryRequest{Id: id})

	if err != nil {
		fmt.Print(err.Error())
	}

	waitc := make(chan struct{})
	go func() {
		for {
			in, err := srv.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				fmt.Print(err.Error())
			}
			fmt.Print(string(in.GetOutput()))
		}
	}()
	<-waitc
}

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	// cert := flag.String("cert", "/opt/pki/tls/certs/cert.pem", "absolute path to the public cert")
	// key := flag.String("cert-key", "/opt/pki/tls/certs/cert.key", "absolute path to the private key of a cert")

	flag.Parse()

	conn, err := grpc.Dial("localhost:8080", opts...)

	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	client := &Client{
		pb.NewJobRunnerServiceClient(conn),
	}

	client.handleArgs(flag.Args())
}
