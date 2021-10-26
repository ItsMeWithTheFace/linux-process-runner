package api

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	pb "github.com/ItsMeWithTheFace/linux-process-runner/api/proto"
	c "github.com/ItsMeWithTheFace/linux-process-runner/core"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// JobRunnerServer implements the server-side gRPC functions.
type JobRunnerServer struct {
	pb.UnimplementedJobRunnerServiceServer
	jr *c.JobRunner
}

// InitializeJobRunnerServer initializes the core functionality to start/stop/get jobs.
func InitializeJobRunnerServer() *JobRunnerServer {
	jr := c.InitializeJobRunner(c.InitializeInMemoryJobStore())
	s := &JobRunnerServer{
		jr: jr,
	}
	return s
}

// GetJobInfo retrieves a job's metadata.
func (s *JobRunnerServer) GetJobInfo(ctx context.Context, req *pb.JobQueryRequest) (*pb.JobInfo, error) {
	job, err := s.jr.GetJob(req.GetId())

	if err != nil {
		return nil, handleError(req.GetId(), err)
	}

	r := &pb.JobInfo{
		Id:        job.Id,
		Command:   job.Cmd.Args[0],
		Arguments: job.Cmd.Args[1:],
		Owner:     job.Owner,
		State:     pb.JobState(job.State),
	}

	if job.Err != nil {
		r.Error = job.Err.Error()
	}

	return r, nil
}

// StartJob creates and runs a job and returns the generated job ID.
func (s *JobRunnerServer) StartJob(ctx context.Context, req *pb.JobStartRequest) (*pb.JobStartOutput, error) {
	id := uuid.NewString()
	cmd := exec.Command(req.GetCommand(), req.GetArguments()...)
	go s.jr.StartJob(id, cmd)

	return &pb.JobStartOutput{Id: id}, nil
}

// StopJob attempts to kill a running job.
func (s *JobRunnerServer) StopJob(ctx context.Context, req *pb.JobStopRequest) (*pb.JobStopOutput, error) {
	err := s.jr.StopJob(req.GetId())

	if err != nil {
		return nil, handleError(req.GetId(), err)
	}

	return &pb.JobStopOutput{}, nil
}

// StreamJobOutput streams a given job's output from their associated log file regardless
// of their state.
func (s *JobRunnerServer) StreamJobOutput(req *pb.JobQueryRequest, srv pb.JobRunnerService_StreamJobOutputServer) error {
	job, err := s.jr.GetJob(req.GetId())

	if err != nil {
		return handleError(req.GetId(), err)
	}

	r, err := job.Output.NewReader()
	defer r.Close()

	if err != nil {
		return handleError(job.Id, err)
	}

	// 16 KB buffer
	buffer := make([]byte, 16*1000)
	for {
		n, err := r.Read(buffer)

		if n == 0 && err == io.EOF && job.State > c.Running {
			return nil
		}

		resp := &pb.JobStreamOutput{Output: buffer[:n]}
		err = srv.Send(resp)
		if err != nil {
			return handleError(job.Id, err)
		}

		job, err = s.jr.GetJob(req.GetId())
		if err != nil {
			handleError(job.Id, err)
		}
	}
}

// handleError customizes returned error messages based on their type.
func handleError(id string, err error) error {
	// TODO: handle other error types and associate error codes with them
	switch err.(type) {
	case *c.ErrNotFound:
		return status.Errorf(
			codes.NotFound,
			fmt.Sprintf("cannot find job with ID: %s Err: %s", id, err.Error()),
		)
	default:
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("error for job: %s Err: %s", id, err.Error()),
		)
	}
}
