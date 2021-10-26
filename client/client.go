package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	pb "github.com/ItsMeWithTheFace/linux-process-runner/api/proto"
	"google.golang.org/grpc"
)

// Client implements the client-side gRPC functions.
type Client struct {
	pb.JobRunnerServiceClient
}

// handleArgs accepts command-line arguments and routes them to the appropriate handler.
func (c *Client) handleArgs(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide one of the following commands: [start, stop, get, stream]")
	}
	// TODO: add some better argument handling
	// TODO: pass in custom context containing TLS credentials
	switch command := args[0]; command {
	case "start":
		return c.handleStartJobCommand(context.Background(), args[1], args[2:])
	case "stop":
		return c.handleStopJobCommand(context.Background(), args[1])
	case "get":
		return c.handleGetJobCommand(context.Background(), args[1])
	case "stream":
		return c.handleStreamJobOutputCommand(context.Background(), args[1])
	default:
		return fmt.Errorf("please provide one of the following commands: [start, stop, get, stream]")
	}
}

// handleStartJobCommand starts the job and returns the job ID.
func (c *Client) handleStartJobCommand(ctx context.Context, command string, args []string) error {
	out, err := c.JobRunnerServiceClient.StartJob(ctx, &pb.JobStartRequest{Command: command, Arguments: args})

	if err != nil {
		return err
	}

	log.Printf("ID: %s", out.GetId())
	return nil
}

// handleStopJobCommand stops the job.
func (c *Client) handleStopJobCommand(ctx context.Context, id string) error {
	_, err := c.JobRunnerServiceClient.StopJob(ctx, &pb.JobStopRequest{Id: id})

	if err != nil {
		return err
	}

	log.Printf("stopped job with ID: %s", id)

	return nil
}

// handleGetJobCommand retrieves a job's metadata.
func (c *Client) handleGetJobCommand(ctx context.Context, id string) error {
	job, err := c.JobRunnerServiceClient.GetJobInfo(ctx, &pb.JobQueryRequest{Id: id})

	if err != nil {
		return err
	}

	log.Println(job)

	return nil
}

// handleStreamJobOutputCommand receives the streamed output of a job and prints it.
func (c *Client) handleStreamJobOutputCommand(ctx context.Context, id string) error {
	srv, err := c.JobRunnerServiceClient.StreamJobOutput(ctx, &pb.JobQueryRequest{Id: id})

	if err != nil {
		fmt.Print(err.Error())
	}

	for {
		in, err := srv.Recv()

		fmt.Print(string(in.GetOutput()))

		if err == io.EOF {
			log.Println("finished reading stream")
			return nil
		}

		if err != nil {
			return fmt.Errorf("stream: %s", err.Error())
		}
	}
}

func main() {
	// TODO: add TLS credentials
	// TODO: add cert and cert-key flags

	flag.Parse()

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())

	if err != nil {
		log.Printf("could not connect to host: %s", err.Error())
		os.Exit(1)
	}

	client := &Client{
		pb.NewJobRunnerServiceClient(conn),
	}

	err = client.handleArgs(flag.Args())
	if err != nil {
		log.Fatalf("error handling command args: %s, err: %s", flag.Args(), err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
