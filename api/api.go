package api

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"os/exec"

	pb "github.com/ItsMeWithTheFace/linux-process-runner/api/proto"
	"github.com/ItsMeWithTheFace/linux-process-runner/auth"
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
	owner, err := getClientID(ctx)

	if err != nil {
		return nil, err
	}

	job := s.jr.CreateJob(id, owner, cmd)

	go s.jr.StartJob(job)

	return &pb.JobStartOutput{Id: id}, nil
}

// StopJob attempts to kill a running job.
func (s *JobRunnerServer) StopJob(ctx context.Context, req *pb.JobStopRequest) (*pb.JobStopOutput, error) {
	job, err := s.jr.GetJob(req.GetId())

	if err != nil {
		return nil, handleError(req.GetId(), err)
	}

	if err = verifyJobOwnership(ctx, job.Owner); err != nil {
		return nil, err
	}

	err = s.jr.StopJob(req.GetId())

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

	if err = verifyJobOwnership(srv.Context(), job.Owner); err != nil {
		return err
	}

	r, err := job.Output.NewReader()
	defer r.Close()

	if err != nil {
		return handleError(job.Id, err)
	}

	// 16 KB buffer
	buffer := make([]byte, 16*1000)

	for {
		select {
		case <-srv.Context().Done():
			return srv.Context().Err()
		default:
			n, err := r.Read(buffer)

			if n == 0 && err == io.EOF && job.State > c.Running {
				return nil
			}

			shouldSkip := err == io.EOF && job.State <= c.Running
			if !shouldSkip {
				resp := &pb.JobStreamOutput{Output: buffer[:n]}
				err = srv.Send(resp)
				if err != nil {
					return handleError(job.Id, err)
				}
			}

			// TODO: change GetJob to return a pointer with privatized fields
			// and Get functions so we avoid querying constantly while maintaining
			// read-only on c.JobInfo
			job, err = s.jr.GetJob(req.GetId())
			if err != nil {
				handleError(job.Id, err)
			}
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

// verifyJobOwnership compares owner IDs to see if they match.
func verifyJobOwnership(ctx context.Context, owner *big.Int) error {
	o, err := getClientID(ctx)

	if err != nil {
		return err
	}

	if o.Cmp(owner) != 0 {
		return status.Errorf(
			codes.PermissionDenied,
			fmt.Sprint("user does not own this job"),
		)
	}

	return nil
}

// getClientID parses a client-side ID so it can be safely tested for job ownership.
func getClientID(ctx context.Context) (*big.Int, error) {
	if o, ok := ctx.Value(auth.ClientIDKey).(*big.Int); ok {
		return o, nil
	}
	return nil, status.Error(codes.Internal, "cannot find owner")
}
