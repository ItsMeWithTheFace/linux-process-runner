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

type JobRunnerServer struct {
	pb.UnimplementedJobRunnerServiceServer
	jr *c.JobRunner
}

func InitializeJobRunnerServer() *JobRunnerServer {
	jr := c.InitializeJobRunner(c.InitializeInMemoryJobStore())
	s := &JobRunnerServer{
		jr: jr,
	}
	return s
}

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

func (s *JobRunnerServer) StartJob(ctx context.Context, req *pb.JobStartRequest) (*pb.JobStartOutput, error) {
	id := uuid.NewString()
	cmd := exec.Command(req.GetCommand(), req.GetArguments()...)
	go s.jr.StartJob(id, cmd)

	return &pb.JobStartOutput{Id: id}, nil
}

func (s *JobRunnerServer) StopJob(ctx context.Context, req *pb.JobStopRequest) (*pb.JobStopOutput, error) {
	err := s.jr.StopJob(req.GetId())

	if err != nil {
		return nil, handleError(req.GetId(), err)
	}

	return &pb.JobStopOutput{}, nil
}

func (s *JobRunnerServer) StreamJobOutput(req *pb.JobQueryRequest, srv pb.JobRunnerService_StreamJobOutputServer) error {
	job, err := s.jr.GetJob(req.GetId())

	if err != nil {
		return status.Errorf(
			codes.NotFound,
			fmt.Sprintf("cannot find job with ID: %s. Err: %s", req.GetId(), err.Error()),
		)
	}

	r, err := job.Output.NewReader()
	defer r.Close()

	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("cannot get output of job with ID: %s. Err: %s", req.GetId(), err.Error()),
		)
	}

	buffer := make([]byte, 16*1000)
	for {
		n, err := r.Read(buffer)

		if n == 0 && err == io.EOF && job.State > c.Running {
			return nil
		}

		resp := &pb.JobStreamOutput{Output: buffer[:n]}
		err = srv.Send(resp)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("error streaming output for job with ID: %s. Err: %s", req.GetId(), err.Error()),
			)
		}

		job, err = s.jr.GetJob(req.GetId())
	}
}

func handleError(id string, err error) error {
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
