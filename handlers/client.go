package handlers

import (
	"context"
	"fmt"
	"io"
	"log"

	pb "github.com/ItsMeWithTheFace/linux-process-runner/api/proto"
)

// Client implements the client-side gRPC functions.
type Client struct {
	pb.JobRunnerServiceClient
}

// HandleArgs accepts command-line arguments and routes them to the appropriate handler.
func (c *Client) HandleArgs(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide one of the following commands: [start, stop, get, stream]")
	}

	// TODO: add some better argument handling
	command := args[0]
	if len(args) < 2 {
		return fmt.Errorf("command %s does not have enough arguments", command)
	}

	switch command {
	case "start":
		return c.HandleStartJobCommand(context.Background(), args[1], args[2:])
	case "stop":
		return c.HandleStopJobCommand(context.Background(), args[1])
	case "get":
		return c.HandleGetJobCommand(context.Background(), args[1])
	case "stream":
		return c.HandleStreamJobOutputCommand(context.Background(), args[1])
	default:
		return fmt.Errorf("please provide one of the following commands: [start, stop, get, stream]")
	}
}

// HandleStartJobCommand starts the job and returns the job ID.
func (c *Client) HandleStartJobCommand(ctx context.Context, command string, args []string) error {
	out, err := c.JobRunnerServiceClient.StartJob(ctx, &pb.JobStartRequest{Command: command, Arguments: args})

	if err != nil {
		return err
	}

	log.Printf("ID: %s", out.GetId())
	return nil
}

// HandleStopJobCommand stops the job.
func (c *Client) HandleStopJobCommand(ctx context.Context, id string) error {
	_, err := c.JobRunnerServiceClient.StopJob(ctx, &pb.JobStopRequest{Id: id})

	if err != nil {
		return err
	}

	log.Printf("stopped job with ID: %s", id)

	return nil
}

// HandleGetJobCommand retrieves a job's metadata.
func (c *Client) HandleGetJobCommand(ctx context.Context, id string) error {
	job, err := c.JobRunnerServiceClient.GetJobInfo(ctx, &pb.JobQueryRequest{Id: id})

	if err != nil {
		return err
	}

	log.Println(job)

	return nil
}

// HandleStreamJobOutputCommand receives the streamed output of a job and prints it.
func (c *Client) HandleStreamJobOutputCommand(ctx context.Context, id string) error {
	srv, err := c.JobRunnerServiceClient.StreamJobOutput(ctx, &pb.JobQueryRequest{Id: id})

	if err != nil {
		fmt.Print(err.Error())
	}

	for {
		in, err := srv.Recv()

		fmt.Print(string(in.GetOutput()))

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return fmt.Errorf("stream: %s", err.Error())
		}
	}
}
