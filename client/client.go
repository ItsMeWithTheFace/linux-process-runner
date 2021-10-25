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
	out, err := c.JobRunnerServiceClient.StartJob(ctx, &pb.JobStartRequest{Command: command, Arguments: args})

	if err != nil {
		fmt.Print(err.Error())
	}

	fmt.Println(out.GetId())
}

func (c *Client) handleStopJobCommand(ctx context.Context, id string) {
	_, err := c.JobRunnerServiceClient.StopJob(ctx, &pb.JobStopRequest{Id: id})

	if err != nil {
		fmt.Print(err.Error())
	}
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

	for {
		in, err := srv.Recv()

		fmt.Print(string(in.GetOutput()))

		if err == io.EOF {
			fmt.Printf("finished reading stream")
			return
		}

		if err != nil {
			fmt.Printf("stream: %s", err.Error())
			return
		}
	}
}

func main() {
	// TODO: add TLS credentials
	// TODO: add cert and cert-key flags

	flag.Parse()

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())

	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	client := &Client{
		pb.NewJobRunnerServiceClient(conn),
	}

	client.handleArgs(flag.Args())
}
